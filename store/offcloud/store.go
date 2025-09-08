package offcloud

import (
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/rodezfranco/stremthru/core"
	"github.com/rodezfranco/stremthru/internal/cache"
	"github.com/rodezfranco/stremthru/internal/request"
	"github.com/rodezfranco/stremthru/store"
)

type StoreClientConfig struct {
	HTTPClient *http.Client
	UserAgent  string
}

type StoreClient struct {
	Name             store.StoreName
	client           *APIClient
	listMagnetsCache cache.Cache[[]store.ListMagnetsDataItem]
}

func NewStoreClient(config *StoreClientConfig) *StoreClient {
	c := &StoreClient{}
	c.client = NewAPIClient(&APIClientConfig{
		HTTPClient: config.HTTPClient,
		UserAgent:  config.UserAgent,
	})
	c.Name = store.StoreNameOffcloud

	c.listMagnetsCache = cache.NewCache[[]store.ListMagnetsDataItem](&cache.CacheConfig{
		Name:     "store:offcloud:listMagnets",
		Lifetime: 5 * time.Minute,
	})

	return c
}

func (c *StoreClient) getCacheKey(params request.Context, key string) string {
	return params.GetAPIKey(c.client.apiKey) + ":" + key
}

func (s *StoreClient) GetName() store.StoreName {
	return s.Name
}

func (s *StoreClient) getMagnetFiles(ctx Ctx, requestId string, server string) ([]store.MagnetFile, string, error) {
	magnetName := ""
	files := []store.MagnetFile{}
	pathByName := map[string]string{}
	if server != "" {
		res, err := s.client.ListCloudDownloadEntries(&ListCloudDownloadEntriesParams{
			Ctx:       ctx,
			RequestId: requestId,
			Server:    server,
		})
		if err != nil {
			return files, "", err
		}
		for _, entry := range res.Data.Entries {
			if filepath.Ext(entry) == ".aria2" {
				continue
			}
			rootFolder, path, _ := strings.Cut(entry, "/")
			pathByName[filepath.Base(entry)] = "/" + path
			if rootFolder != "" {
				magnetName = rootFolder
			}
		}
	}

	res, err := s.client.ExploreCloudDownload(&ExploreCloudDownloadParams{
		Ctx:       ctx,
		RequestId: requestId,
	})
	if err != nil {
		return files, "", err
	}
	for _, link := range res.Data {
		info, err := link.parse()
		if err != nil {
			return nil, "", err
		}
		if filepath.Ext(info.fileName) == ".aria2" {
			continue
		}
		size := int64(-1)
		// // too expensive, should enable for non-eublic deployments later
		// if size_res, err := s.client.GetFileSize(&GetFileSizeParams{Ctx: ctx, Link: string(link)}); err == nil {
		// 	size = size_res.Data
		// }
		file := store.MagnetFile{
			Idx:  info.fileIdx,
			Link: string(link),
			Name: info.fileName,
			Size: size,
		}
		if path, ok := pathByName[file.Name]; ok {
			file.Path = path
		}
		files = append(files, file)
	}
	return files, magnetName, nil
}

func (s *StoreClient) AddMagnet(params *store.AddMagnetParams) (*store.AddMagnetData, error) {
	magnet, err := core.ParseMagnetLink(params.Magnet)
	if err != nil {
		return nil, err
	}

	cloudDownload, err := s.findCloudDownload(params.Ctx, magnet.Hash)
	if err != nil {
		return nil, err
	}

	data := &store.AddMagnetData{
		Hash:   magnet.Hash,
		Magnet: magnet.Link,
		Size:   -1,
		Files:  []store.MagnetFile{},
	}

	server := ""

	if cloudDownload != nil {
		data.Id = cloudDownload.RequestId
		data.Name = cloudDownload.FileName
		data.Status = getMagnetStatus(cloudDownload.Status)
		data.AddedAt = cloudDownload.CreatedOn
		data.Size = cloudDownload.FileSize

		server = cloudDownload.Server
	} else {
		res, err := s.client.AddCloudDownload(&AddCloudDownloadParams{
			Ctx: params.Ctx,
			URL: magnet.RawLink,
		})
		if err != nil {
			return nil, err
		}

		data.Id = res.Data.RequestId
		data.Name = res.Data.FileName
		data.Status = getMagnetStatus(res.Data.Status)
		data.AddedAt = res.Data.CreatedOn

		server = res.Data.GetServer()

		s.listMagnetsCache.Remove(s.getCacheKey(params, ""))
	}

	if data.Status == store.MagnetStatusDownloaded {
		files, name, err := s.getMagnetFiles(params.Ctx, data.Id, server)
		if err != nil {
			return nil, err
		}
		data.Name = name
		data.Files = files
	}

	return data, nil
}

func (s *StoreClient) CheckMagnet(params *store.CheckMagnetParams) (*store.CheckMagnetData, error) {
	hashes := []string{}
	magnetByHash := map[string]core.MagnetLink{}
	for _, magnet := range params.Magnets {
		if m, err := core.ParseMagnetLink(magnet); err == nil {
			hashes = append(hashes, m.Hash)
			magnetByHash[m.Hash] = m
		}
	}
	res, err := s.client.CheckCache(&CheckCacheParams{
		Ctx:    params.Ctx,
		Hashes: hashes,
	})
	if err != nil {
		return nil, err
	}
	data := &store.CheckMagnetData{
		Items: []store.CheckMagnetDataItem{},
	}
	cachedByHash := map[string]bool{}
	for _, hash := range res.Data.CachedItems {
		cachedByHash[hash] = true
	}
	for _, hash := range hashes {
		m := magnetByHash[hash]
		item := store.CheckMagnetDataItem{
			Hash:   m.Hash,
			Magnet: m.Link,
			Status: store.MagnetStatusUnknown,
			Files:  []store.MagnetFile{},
		}
		if _, ok := cachedByHash[m.Hash]; ok {
			item.Status = store.MagnetStatusCached
		}
		data.Items = append(data.Items, item)
	}
	return data, nil
}

func (s *StoreClient) GenerateLink(params *store.GenerateLinkParams) (*store.GenerateLinkData, error) {
	data := &store.GenerateLinkData{Link: params.Link}
	return data, nil
}

func getMagnetStatus(status CloudDownloadStatus) store.MagnetStatus {
	switch status {
	case CloudDownloadStatusCreated:
		return store.MagnetStatusQueued
	case CloudDownloadStatusDownloading:
		return store.MagnetStatusDownloading
	case CloudDownloadStatusDownloaded:
		return store.MagnetStatusDownloaded
	case CloudDownloadStatusError:
		return store.MagnetStatusFailed
	default:
		return store.MagnetStatusUnknown
	}
}

func (s *StoreClient) GetMagnet(params *store.GetMagnetParams) (*store.GetMagnetData, error) {
	res, err := s.client.GetCloudDownloadStatus(&GetCloudDownloadStatusParams{
		Ctx:       params.Ctx,
		RequestId: params.Id,
	})
	if err != nil {
		return nil, err
	}
	magnet := res.Data.Status
	data := &store.GetMagnetData{
		Id:      params.Id,
		Name:    magnet.FileName,
		Hash:    "",
		Size:    magnet.FileSize,
		Status:  getMagnetStatus(magnet.Status),
		Files:   []store.MagnetFile{},
		AddedAt: time.Unix(0, 0),
	}
	if data.Status == store.MagnetStatusDownloaded {
		files, name, err := s.getMagnetFiles(params.Ctx, data.Id, res.Data.Status.Server)
		if err != nil {
			return nil, err
		}
		data.Name = name
		data.Files = files
	}
	return data, nil
}

func (s *StoreClient) GetUser(params *store.GetUserParams) (*store.User, error) {
	email_res, err := s.client.GetUserEmail(&GetUserEmailParams{
		Ctx: params.Ctx,
	})
	if err != nil {
		return nil, err
	}
	data := &store.User{
		Id:                 email_res.Data.UserId,
		Email:              email_res.Data.Email,
		SubscriptionStatus: store.UserSubscriptionStatusTrial,
	}
	stats_res, err := s.client.GetAccountStats(&GetAccountStatsParams{
		Ctx: params.Ctx,
	})
	if err != nil {
		return nil, err
	}
	if stats_res.Data.ExpirationDate.After(time.Now()) {
		data.SubscriptionStatus = store.UserSubscriptionStatusPremium
	}
	return data, nil
}

func (s *StoreClient) findCloudDownload(ctx Ctx, hash string) (*ListCloudDownloadsDataItem, error) {
	res, err := s.client.ListCloudDownloads(&ListCloudDownloadsParams{
		Ctx: ctx,
	})
	if err != nil {
		return nil, err
	}
	for i := range res.Data.History {
		item := &res.Data.History[i]
		if strings.Contains(item.OriginalLink, hash) {
			return item, nil
		}
	}
	return nil, nil
}

func (s *StoreClient) ListMagnets(params *store.ListMagnetsParams) (*store.ListMagnetsData, error) {
	lm := []store.ListMagnetsDataItem{}

	if !s.listMagnetsCache.Get(s.getCacheKey(params, ""), &lm) {
		items := []store.ListMagnetsDataItem{}
		page := 0
		pageSize := -1

		for {
			res, err := s.client.ListCloudDownloads(&ListCloudDownloadsParams{
				Ctx:  params.Ctx,
				Page: page,
			})
			if err != nil {
				return nil, err
			}
			for _, m := range res.Data.History {
				magnet, err := core.ParseMagnetLink(m.OriginalLink)
				if err != nil {
					continue
				}
				item := store.ListMagnetsDataItem{
					Id:      m.RequestId,
					Hash:    magnet.Hash,
					Name:    m.FileName,
					Size:    m.FileSize,
					Status:  getMagnetStatus(m.Status),
					AddedAt: m.CreatedOn,
				}
				items = append(items, item)
			}

			if res.Data.IsEnd {
				break
			}
			if pageSize == -1 {
				pageSize = len(res.Data.History)
				log.Info("found page size", "pageSize", pageSize)
			}
			page += 1
		}

		lm = items
		s.listMagnetsCache.Add(s.getCacheKey(params, ""), items)
	}

	totalItems := len(lm)
	startIdx := min(params.Offset, totalItems)
	endIdx := min(startIdx+params.Limit, totalItems)
	items := lm[startIdx:endIdx]

	data := &store.ListMagnetsData{
		Items:      items,
		TotalItems: totalItems,
	}

	return data, nil
}

func (s *StoreClient) RemoveMagnet(params *store.RemoveMagnetParams) (*store.RemoveMagnetData, error) {
	_, err := s.client.RemoveCloudDownload(&RemoveCloudDownloadParams{
		Ctx:       params.Ctx,
		RequestId: params.Id,
	})
	if err != nil {
		return nil, err
	}

	s.listMagnetsCache.Remove(s.getCacheKey(params, ""))

	data := &store.RemoveMagnetData{Id: params.Id}
	return data, nil
}
