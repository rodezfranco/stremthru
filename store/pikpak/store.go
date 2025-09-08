package pikpak

import (
	"net/http"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/MunifTanjim/stremthru/core"
	"github.com/MunifTanjim/stremthru/internal/buddy"
	"github.com/MunifTanjim/stremthru/internal/cache"
	"github.com/MunifTanjim/stremthru/store"
)

func toSize(sizeStr string) int64 {
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		size = -1
	}
	return int64(size)
}

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
	c.Name = store.StoreNamePikPak

	c.listMagnetsCache = func() cache.Cache[[]store.ListMagnetsDataItem] {
		return cache.NewCache[[]store.ListMagnetsDataItem](&cache.CacheConfig{
			Name:     "store:pikpak:listMagnets",
			Lifetime: 5 * time.Minute,
		})
	}()

	return c
}

func (s *StoreClient) getCacheKey(ctx Ctx, key string) string {
	return ctx.GetDeviceId() + ":" + key
}

func (s *StoreClient) GetName() store.StoreName {
	return s.Name
}

func (s *StoreClient) getRecentTask(ctx Ctx, taskId string) (*Task, error) {
	res, err := s.client.ListTasks(&ListTasksParams{
		Ctx:   ctx,
		Limit: 200,
		Filters: map[string]map[string]any{
			"phase": {
				"in": FilePhaseRunning + "," + FilePhaseError + "," + FilePhaseComplete,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	for i := range res.Data.Tasks {
		t := &res.Data.Tasks[i]
		if t.Id == taskId {
			return t, nil
		}
	}
	error := core.NewStoreError("task not found: " + string(taskId))
	error.StoreName = string(s.GetName())
	return nil, error
}

func (s *StoreClient) waitForTaskComplete(ctx Ctx, taskId string, maxRetry int, retryInterval time.Duration) (*Task, error) {
	t, err := s.getRecentTask(ctx, taskId)
	if err != nil {
		return nil, err
	}
	retry := 0
	for (t.Phase != FilePhaseComplete && t.Phase != FilePhaseError) && retry < maxRetry {
		time.Sleep(retryInterval)
		task, err := s.getRecentTask(ctx, t.Id)
		if err != nil {
			return t, err
		}
		t = task
		retry++
	}
	if t.Phase != FilePhaseComplete {
		error := core.NewStoreError("task failed to reach phase: " + string(FilePhaseComplete))
		error.StoreName = string(s.GetName())
		return t, error
	}
	return t, nil
}

func (s *StoreClient) getFileByMagnetHash(ctx Ctx, hash string) (*File, error) {
	myPackFolder, err := s.getMyPackFolder(ctx)
	if err != nil {
		return nil, err
	}

	res, err := s.client.ListFiles(&ListFilesParams{
		Ctx:      ctx,
		Limit:    500,
		ParentId: myPackFolder.Id,
		Filters: map[string]map[string]any{
			"trashed": {"eq": false},
			"phase":   {"eq": FilePhaseComplete},
		},
	})
	if err != nil {
		return nil, err
	}
	for i := range res.Data.Files {
		f := &res.Data.Files[i]
		if strings.Contains(f.Params.URL, hash) {
			return f, nil
		}
	}
	return nil, nil
}

func (s *StoreClient) AddMagnet(params *store.AddMagnetParams) (*store.AddMagnetData, error) {
	magnet, err := core.ParseMagnetLink(params.Magnet)
	if err != nil {
		return nil, err
	}
	ctx := Ctx{Ctx: params.Ctx}

	file, err := s.getFileByMagnetHash(ctx, magnet.Hash)
	if err != nil {
		return nil, err
	}

	data := &store.AddMagnetData{
		Hash:    magnet.Hash,
		Magnet:  magnet.Link,
		Name:    "",
		Size:    -1,
		Status:  store.MagnetStatusQueued,
		Files:   []store.MagnetFile{},
		AddedAt: time.Now(),
	}

	if file != nil {
		data.Id = file.Id

		mRes, err := s.GetMagnet(&store.GetMagnetParams{
			Ctx: ctx.Ctx,
			Id:  data.Id,
		})
		if err != nil {
			return nil, err
		}
		data.Name = mRes.Name
		data.Status = mRes.Status
		data.Files = mRes.Files
		data.AddedAt = mRes.AddedAt
		return data, nil
	}

	res, err := s.client.AddFile(&AddFileParams{
		Ctx: ctx,
		URL: AddFileParamsURL{
			URL: magnet.RawLink,
		},
	})
	if err != nil {
		return nil, err
	}

	s.listMagnetsCache.Remove(s.getCacheKey(ctx, ""))

	data.Id = res.Data.Task.FileId
	if task, err := s.waitForTaskComplete(ctx, res.Data.Task.Id, 3, 5*time.Second); task != nil {
		if err != nil {
			log.Error("error waiting for task complete", "error", err)
		}
		if task.Phase == FilePhaseComplete {
			mRes, err := s.GetMagnet(&store.GetMagnetParams{
				Ctx: ctx.Ctx,
				Id:  data.Id,
			})
			if err != nil {
				return nil, err
			}
			data.Name = mRes.Name
			data.Status = mRes.Status
			data.Files = mRes.Files
			data.AddedAt = mRes.AddedAt
		} else if task.Phase == FilePhaseError {
			data.Status = store.MagnetStatusFailed
		}
	}
	return data, nil
}

func (s *StoreClient) CheckMagnet(params *store.CheckMagnetParams) (*store.CheckMagnetData, error) {
	hashes := []string{}
	for _, m := range params.Magnets {
		magnet, err := core.ParseMagnetLink(m)
		if err != nil {
			return nil, err
		}
		hashes = append(hashes, magnet.Hash)
	}

	data, err := buddy.CheckMagnet(s, hashes, params.GetAPIKey(s.client.apiKey), params.ClientIP, params.SId)
	if err != nil {
		return nil, err
	}
	return data, nil
}

type LockedFileLink string

const lockedFileLinkPrefix = "stremthru://store/pikpak/"

func (l LockedFileLink) encodeData(rootFileId, fileId string) string {
	return core.Base64Encode(rootFileId + ":" + fileId)
}

func (l LockedFileLink) decodeData(encoded string) (rootFileId, fileId string, err error) {
	decoded, err := core.Base64Decode(encoded)
	if err != nil {
		return "", "", err
	}
	rootFileId, fileId, found := strings.Cut(decoded, ":")
	if !found {
		return "", "", err
	}
	return rootFileId, fileId, nil
}

func (l LockedFileLink) create(rootFileId, fileId string) string {
	return lockedFileLinkPrefix + l.encodeData(rootFileId, fileId)
}

func (l LockedFileLink) parse() (rootFileId, fileId string, err error) {
	encoded := strings.TrimPrefix(string(l), lockedFileLinkPrefix)
	return l.decodeData(encoded)
}

func (s *StoreClient) GenerateLink(params *store.GenerateLinkParams) (*store.GenerateLinkData, error) {
	_, fileId, err := LockedFileLink(params.Link).parse()
	if err != nil {
		error := core.NewAPIError("invalid link")
		error.StoreName = string(s.GetName())
		error.StatusCode = http.StatusBadRequest
		error.Cause = err
		return nil, error
	}
	ctx := Ctx{Ctx: params.Ctx}
	res, err := s.client.GetFile(&GetFileParams{
		Ctx:    ctx,
		FileId: fileId,
	})
	if err != nil {
		return nil, err
	}
	if len(res.Data.Medias) == 0 {
		err := core.NewStoreError("file not found")
		err.StoreName = string(s.GetName())
		err.StatusCode = http.StatusNotFound
		return nil, err
	}
	data := &store.GenerateLinkData{
		Link: res.Data.Medias[0].Link.URL,
	}
	return data, nil
}

func (c *StoreClient) listFilesFlat(ctx Ctx, folderId string, result []store.MagnetFile, parent *store.MagnetFile, rootFolderId string) ([]store.MagnetFile, error) {
	if result == nil {
		result = []store.MagnetFile{}
	}

	params := &ListFilesParams{
		Ctx:      ctx,
		ParentId: folderId,
		Filters: map[string]map[string]any{
			"trashed": {"eq": false},
			"phase":   {"eq": FilePhaseComplete},
		},
	}
	lfRes, err := c.client.ListFiles(params)
	if err != nil {
		return nil, err
	}

	for _, f := range lfRes.Data.Files {
		file := &store.MagnetFile{
			Idx:  -1, // order is non-deterministic
			Link: LockedFileLink("").create(rootFolderId, f.Id),
			Name: f.Name,
			Path: "/" + f.Name,
			Size: toSize(f.Size),
		}

		if parent != nil {
			file.Path = path.Join(parent.Path, file.Name)
		}

		if f.Kind == FileKindFolder {
			result, err = c.listFilesFlat(ctx, f.Id, result, file, rootFolderId)
			if err != nil {
				return nil, err
			}
		} else {
			result = append(result, *file)
		}
	}

	return result, nil
}

func (s *StoreClient) GetMagnet(params *store.GetMagnetParams) (*store.GetMagnetData, error) {
	ctx := Ctx{Ctx: params.Ctx}
	res, err := s.client.GetFile(&GetFileParams{
		Ctx:    ctx,
		FileId: params.Id,
	})
	if err != nil {
		return nil, err
	}
	magnet, err := core.ParseMagnetLink(res.Data.Params.URL)
	if err != nil {
		return nil, err
	}
	addedAt, err := time.Parse(time.RFC3339, res.Data.CreatedTime)
	if err != nil {
		addedAt = time.Unix(0, 0)
	}
	data := &store.GetMagnetData{
		Id:      res.Data.Id,
		Name:    res.Data.Name,
		Hash:    magnet.Hash,
		Size:    -1,
		Status:  store.MagnetStatusDownloading,
		Files:   []store.MagnetFile{},
		AddedAt: addedAt,
	}
	if res.Data.Phase == FilePhaseComplete {
		data.Status = store.MagnetStatusDownloaded
		if res.Data.Kind == FileKindFolder {
			files, err := s.listFilesFlat(ctx, data.Id, nil, nil, data.Id)
			if err != nil {
				return nil, err
			}
			data.Files = files
			data.Size = 0
			for i := range files {
				data.Size += files[i].Size
			}
		} else {
			data.Files = append(data.Files, store.MagnetFile{
				Idx:  -1,
				Link: LockedFileLink("").create(data.Id, data.Id),
				Name: data.Name,
				Path: "/" + data.Name,
				Size: toSize(res.Data.Size),
			})
		}
	}
	return data, nil
}

func (s *StoreClient) GetUser(params *store.GetUserParams) (*store.User, error) {
	res, err := s.client.GetUser(&GetUserParams{
		Ctx: Ctx{Ctx: params.Ctx},
	})
	if err != nil {
		return nil, err
	}
	vipRes, err := s.client.GetVIPInfo(&GetVIPInfoParams{
		Ctx: Ctx{Ctx: params.Ctx},
	})
	if err != nil {
		return nil, err
	}
	data := &store.User{
		Id:                 res.Data.Sub,
		Email:              res.Data.Email,
		SubscriptionStatus: store.UserSubscriptionStatusTrial,
	}
	if vipRes.Data.Type == VIPTypePlatinum {
		data.SubscriptionStatus = store.UserSubscriptionStatusPremium
	}
	return data, nil
}

func (s *StoreClient) getMyPackFolder(ctx Ctx) (*File, error) {
	res, err := s.client.ListFiles(&ListFilesParams{
		Ctx: ctx,
		Filters: map[string]map[string]any{
			"trashed": {"eq": false},
			"phase":   {"eq": FilePhaseComplete},
		},
	})
	if err != nil {
		return nil, err
	}
	for i := range res.Data.Files {
		f := &res.Data.Files[i]
		if f.Name == "My Pack" {
			return f, nil
		}
	}
	err = core.NewAPIError("'My Pack' folder missing")
	return nil, err
}

func (s *StoreClient) ListMagnets(params *store.ListMagnetsParams) (*store.ListMagnetsData, error) {
	ctx := Ctx{Ctx: params.Ctx}

	lm := []store.ListMagnetsDataItem{}

	if !s.listMagnetsCache.Get(s.getCacheKey(ctx, ""), &lm) {
		items := []store.ListMagnetsDataItem{}
		pageToken := ""
		for {
			myPackFolder, err := s.getMyPackFolder(ctx)
			if err != nil {
				return nil, err
			}
			res, err := s.client.ListFiles(&ListFilesParams{
				Ctx:      Ctx{Ctx: params.Ctx},
				Limit:    500,
				ParentId: myPackFolder.Id,
				Filters: map[string]map[string]any{
					"trashed": {"eq": false},
					"phase":   {"eq": FilePhaseComplete},
				},
				PageToken: pageToken,
			})
			if err != nil {
				return nil, err
			}

			for i := range res.Data.Files {
				f := &res.Data.Files[i]
				addedAt, err := time.Parse(time.RFC3339, f.CreatedTime)
				if err != nil {
					addedAt = time.Unix(0, 0)
				}
				if !strings.HasPrefix(f.Params.URL, "magnet:") {
					continue
				}
				magnet, err := core.ParseMagnetLink(f.Params.URL)
				if err != nil {
					continue
				}
				item := store.ListMagnetsDataItem{
					Id:      f.Id,
					Name:    f.Name,
					Hash:    magnet.Hash,
					Size:    toSize(f.Size),
					Status:  store.MagnetStatusDownloading,
					AddedAt: addedAt,
				}
				if f.Phase == FilePhaseComplete {
					item.Status = store.MagnetStatusDownloaded
				}
				items = append(items, item)
			}

			pageToken = res.Data.NextPageToken
			if pageToken == "" {
				break
			}
		}

		slices.SortFunc(items, func(a, b store.ListMagnetsDataItem) int {
			return b.AddedAt.Compare(a.AddedAt)
		})

		lm = items
		s.listMagnetsCache.Add(s.getCacheKey(ctx, ""), items)
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
	ctx := Ctx{Ctx: params.Ctx}
	_, err := s.client.Trash(&TrashParams{
		Ctx: ctx,
		Ids: []string{params.Id},
	})
	if err != nil {
		return nil, err
	}

	s.listMagnetsCache.Remove(s.getCacheKey(ctx, ""))

	data := &store.RemoveMagnetData{
		Id: params.Id,
	}
	return data, nil
}
