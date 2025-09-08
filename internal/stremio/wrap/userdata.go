package stremio_wrap

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rodezfranco/stremthru/internal/cache"
	"github.com/rodezfranco/stremthru/internal/config"
	"github.com/rodezfranco/stremthru/internal/context"
	"github.com/rodezfranco/stremthru/internal/server"
	"github.com/rodezfranco/stremthru/internal/shared"
	stremio_addon "github.com/rodezfranco/stremthru/internal/stremio/addon"
	stremio_transformer "github.com/rodezfranco/stremthru/internal/stremio/transformer"
	stremio_userdata "github.com/rodezfranco/stremthru/internal/stremio/userdata"
	"github.com/rodezfranco/stremthru/store"
	"github.com/rodezfranco/stremthru/stremio"
)

var upstreamResolverCache = cache.NewCache[upstreamsResolver](&cache.CacheConfig{
	Name:     "stremio:wrap:upstreamResolver",
	Lifetime: 24 * time.Hour,
})

type upstreamsResolverEntry struct {
	Prefix  string
	Indices []int
}

type upstreamsResolver map[string][]upstreamsResolverEntry

func (usr upstreamsResolver) resolve(ud UserData, rName stremio.ResourceName, rType string, id string) []UserDataUpstream {
	upstreams := []UserDataUpstream{}
	key := string(rName) + ":" + rType
	if _, found := usr[key]; !found {
		return upstreams
	}
	for _, entry := range usr[key] {
		if strings.HasPrefix(id, entry.Prefix) {
			for _, idx := range entry.Indices {
				upstreams = append(upstreams, ud.Upstreams[idx])
			}
			break
		}
	}
	return upstreams
}

type UserDataUpstream struct {
	URL              string                                  `json:"u"`
	baseUrl          *url.URL                                `json:"-"`
	ExtractorId      string                                  `json:"e,omitempty"`
	extractor        stremio_transformer.StreamExtractorBlob `json:"-"`
	NoContentProxy   bool                                    `json:"ncp,omitempty"`
	ReconfigureStore bool                                    `json:"rs,omitempty"`
}

type UserData struct {
	Upstreams   []UserDataUpstream `json:"upstreams"`
	ManifestURL string             `json:"manifest_url,omitempty"`

	IncludeTorz bool `json:"torz,omitempty"`

	stremio_userdata.UserDataStores
	StoreName  string `json:"store,omitempty"`
	StoreToken string `json:"token,omitempty"`

	CachedOnly bool `json:"cached,omitempty"`

	TemplateId string                                 `json:"template,omitempty"`
	template   stremio_transformer.StreamTemplateBlob `json:"-"`

	Sort string `json:"sort,omitempty"`

	RPDBAPIKey string `json:"rpdb_akey,omitempty"`

	encoded   string             `json:"-"` // correctly configured
	manifests []stremio.Manifest `json:"-"`
	resolver  upstreamsResolver  `json:"-"`
}

var udManager = stremio_userdata.NewManager[UserData](&stremio_userdata.ManagerConfig{
	AddonName: "wrap",
})

func (ud UserData) HasRequiredValues() bool {
	if len(ud.Upstreams) == 0 {
		return false
	}
	for i := range ud.Upstreams {
		if ud.Upstreams[i].URL == "" {
			return false
		}
	}
	if len(ud.Stores) == 0 {
		return false
	}
	for i := range ud.Stores {
		s := &ud.Stores[i]
		if s.Code.IsStremThru() && len(ud.Stores) > 1 {
			return false
		}
		if s.Token == "" {
			return false
		}
	}
	return true
}

func (ud *UserData) GetEncoded() string {
	return ud.encoded
}

func (ud *UserData) SetEncoded(encoded string) {
	ud.encoded = encoded
}

func (ud *UserData) Ptr() *UserData {
	return ud
}

type userDataError struct {
	upstreamUrl []string
	store       []string
	token       []string
}

func (uderr *userDataError) Error() string {
	var str strings.Builder
	hasSome := false
	for i, err := range uderr.upstreamUrl {
		if err != "" {
			if hasSome {
				str.WriteString(", ")
				hasSome = false
			}

			str.WriteString("upstream_url[" + strconv.Itoa(i) + "]: ")
			str.WriteString(err)
			hasSome = true
		}
	}
	for i, err := range uderr.store {
		if err == "" {
			continue
		}
		if hasSome {
			str.WriteString(", ")
			hasSome = false
		}
		str.WriteString("store[" + strconv.Itoa(i) + "]: ")
		str.WriteString(err)
		hasSome = true
	}
	for i, err := range uderr.token {
		if err == "" {
			continue
		}
		if hasSome {
			str.WriteString(", ")
			hasSome = false
		}
		str.WriteString("token[" + strconv.Itoa(i) + "]: ")
		str.WriteString(err)
		hasSome = true

	}
	return str.String()
}

func (ud *UserData) GetRequestContext(r *http.Request) (*context.StoreContext, error) {
	rCtx := server.GetReqCtx(r)
	ctx := &context.StoreContext{
		Log: rCtx.Log,
	}

	upstreamUrlErrors := []string{}
	hasUpstreamUrlErrors := false
	for i := range ud.Upstreams {
		up := &ud.Upstreams[i]
		if up.baseUrl == nil {
			upstreamUrlErrors = append(upstreamUrlErrors, "Invalid Manifest URL")
			hasUpstreamUrlErrors = true
		} else {
			upstreamUrlErrors = append(upstreamUrlErrors, "")
		}
	}
	if hasUpstreamUrlErrors {
		return ctx, &userDataError{upstreamUrl: upstreamUrlErrors}
	}

	if err, errField := ud.UserDataStores.Prepare(ctx); err != nil {
		switch errField {
		case "store":
			return ctx, &userDataError{store: []string{err.Error()}}
		case "token":
			return ctx, &userDataError{token: []string{err.Error()}}
		default:
			return ctx, &userDataError{store: []string{err.Error()}}
		}
	}

	ctx.ClientIP = shared.GetClientIP(r, ctx)

	return ctx, nil
}

func (ud UserData) getUpstreamManifests(ctx *context.StoreContext) ([]stremio.Manifest, []error) {
	if ud.manifests == nil {
		var wg sync.WaitGroup

		manifests := make([]stremio.Manifest, len(ud.Upstreams))
		errs := make([]error, len(ud.Upstreams))
		hasError := false
		for i := range ud.Upstreams {
			up := &ud.Upstreams[i]
			wg.Add(1)
			go func() {
				defer wg.Done()
				res, err := addon.GetManifest(&stremio_addon.GetManifestParams{BaseURL: up.baseUrl, ClientIP: ctx.ClientIP})
				manifests[i] = res.Data
				errs[i] = err
				if err != nil {
					hasError = true
				}
			}()
		}

		wg.Wait()

		if hasError {
			return manifests, errs
		}

		ud.manifests = manifests
	}

	return ud.manifests, nil
}

func (ud UserData) getUpstreamsResolver(ctx *context.StoreContext) (upstreamsResolver, error) {
	eud := ud.GetEncoded()

	if ud.resolver == nil {
		if upstreamResolverCache.Get(eud, &ud.resolver) {
			return ud.resolver, nil
		}

		manifests, errs := ud.getUpstreamManifests(ctx)
		if errs != nil {
			return nil, errors.Join(errs...)
		}

		resolver := upstreamsResolver{}
		entryIdxMap := map[string]int{}
		for mIdx := range manifests {
			m := &manifests[mIdx]
			for _, r := range m.Resources {
				if r.Name == stremio.ResourceNameAddonCatalog || r.Name == stremio.ResourceNameCatalog {
					continue
				}

				idPrefixes := getManifestResourceIdPrefixes(m, r)
				for _, rType := range getManifestResourceTypes(m, r) {
					key := string(r.Name) + ":" + string(rType)
					if _, found := resolver[key]; !found {
						resolver[key] = []upstreamsResolverEntry{}
					}
					for _, idPrefix := range idPrefixes {
						idPrefixKey := key + ":" + idPrefix
						if idx, found := entryIdxMap[idPrefixKey]; found {
							resolver[key][idx].Indices = append(resolver[key][idx].Indices, mIdx)
						} else {
							resolver[key] = append(resolver[key], upstreamsResolverEntry{
								Prefix:  idPrefix,
								Indices: []int{mIdx},
							})
							entryIdxMap[idPrefixKey] = len(resolver[key]) - 1
						}
					}
				}
			}
		}

		err := upstreamResolverCache.Add(eud, resolver)
		if err != nil {
			return nil, err
		}

		ud.resolver = resolver
	}

	return ud.resolver, nil
}

func (ud UserData) getUpstreams(ctx *context.StoreContext, rName stremio.ResourceName, rType, id string) ([]UserDataUpstream, error) {
	switch rName {
	case stremio.ResourceNameAddonCatalog:
		return []UserDataUpstream{}, nil
	case stremio.ResourceNameCatalog:
		return []UserDataUpstream{}, nil
	default:
		upstreamsCount := len(ud.Upstreams)
		if upstreamsCount == 1 {
			return ud.Upstreams, nil
		}

		if IsPublicInstance {
			if rName == stremio.ResourceNameMeta || rName == stremio.ResourceNameSubtitles {
				if upstreamsCount > 1 {
					return []UserDataUpstream{}, nil
				}
			}
			return ud.Upstreams, nil
		}

		resolver, err := ud.getUpstreamsResolver(ctx)
		if err != nil {
			return nil, err
		}
		return resolver.resolve(ud, rName, rType, id), nil
	}
}

func getUserData(r *http.Request) (*UserData, error) {
	data := &UserData{}
	data.SetEncoded(r.PathValue("userData"))

	log := server.GetReqCtx(r).Log

	if IsMethod(r, http.MethodGet) || IsMethod(r, http.MethodHead) {
		if data.GetEncoded() == "" {
			return data, nil
		}

		if err := udManager.Resolve(data); err != nil {
			return nil, err
		}
		if data.encoded == "" {
			return data, nil
		}

		shouldResync := false
		if data.StoreToken != "" {
			data.Stores = []stremio_userdata.Store{
				{
					Code:  stremio_userdata.StoreCode(store.StoreName(data.StoreName).Code()),
					Token: data.StoreToken,
				},
			}
			data.StoreName = ""
			data.StoreToken = ""
			shouldResync = true
		}

		if data.ManifestURL != "" {
			data.Upstreams = []UserDataUpstream{
				{
					URL: data.ManifestURL,
				},
			}
			data.ManifestURL = ""
			shouldResync = true
		}

		if shouldResync {
			if err := udManager.Sync(data); err != nil {
				return nil, err
			}
		}

		for i := range data.Upstreams {
			up := &data.Upstreams[i]

			if up.ExtractorId != "" {
				if config.IsPublicInstance {
					up.ExtractorId = getNewTransformerExtractorId(up.ExtractorId)
				}

				if extractor, err := getExtractor(up.ExtractorId); err != nil {
					LogError(r, fmt.Sprintf("failed to fetch extractor(%s)", up.ExtractorId), err)
				} else {
					up.extractor = extractor
				}
			}
		}

		if data.TemplateId != "" {
			if config.IsPublicInstance && !strings.HasPrefix(data.TemplateId, BUILTIN_TRANSFORMER_ENTITY_ID_PREFIX) {
				data.TemplateId = BUILTIN_TRANSFORMER_ENTITY_ID_PREFIX + data.TemplateId
			}

			if template, err := getTemplate(data.TemplateId); err != nil {
				LogError(r, fmt.Sprintf("failed to fetch template(%s)", data.TemplateId), err)
			} else {
				data.template = template
			}
		}
	}

	if IsMethod(r, http.MethodPost) {
		err := r.ParseForm()
		if err != nil {
			return nil, err
		}

		upstreams_length := 1
		if v := r.Form.Get("upstreams_length"); v != "" {
			upstreams_length, err = strconv.Atoi(v)
			if err != nil {
				return nil, err
			}
		}

		data.IncludeTorz = r.Form.Get("torz") == "on"
		data.Sort = r.Form.Get("sort")
		data.RPDBAPIKey = r.Form.Get("rpdb_akey")

		data.TemplateId = r.Form.Get("transformer.template_id")
		data.template = stremio_transformer.StreamTemplateBlob{
			Name:        r.Form.Get("transformer.template.name"),
			Description: r.Form.Get("transformer.template.description"),
		}

		for idx := range upstreams_length {
			upURL, err := stremio_addon.NormalizeManifestURL(r.Form.Get("upstreams[" + strconv.Itoa(idx) + "].url"))
			if err != nil {
				log.Error("failed to normalize manifest url", "error", err)
			}
			extractorId := r.Form.Get("upstreams[" + strconv.Itoa(idx) + "].transformer.extractor_id")
			up := UserDataUpstream{
				URL:         upURL,
				ExtractorId: extractorId,
			}
			extractor := r.Form.Get("upstreams[" + strconv.Itoa(idx) + "].transformer.extractor")
			if extractor != "" {
				up.extractor = stremio_transformer.StreamExtractorBlob(extractor)
			}
			if upURL != "" || extractorId != "" || extractor != "" {
				up.NoContentProxy = r.Form.Get("upstreams["+strconv.Itoa(idx)+"].no_content_proxy") == "on"
				up.ReconfigureStore = r.Form.Get("upstreams["+strconv.Itoa(idx)+"].reconfigure_store") == "on"
				data.Upstreams = append(data.Upstreams, up)
			}
		}

		stores_length := 1
		if v := r.Form.Get("stores_length"); v != "" {
			stores_length, err = strconv.Atoi(v)
			if err != nil {
				return nil, err
			}
		}

		for idx := range stores_length {
			code := r.Form.Get("stores[" + strconv.Itoa(idx) + "].code")
			token := r.Form.Get("stores[" + strconv.Itoa(idx) + "].token")
			if code == "" {
				data.Stores = []stremio_userdata.Store{
					{
						Code:  stremio_userdata.StoreCode(code),
						Token: token,
					},
				}
				break
			} else {
				data.Stores = append(data.Stores, stremio_userdata.Store{
					Code:  stremio_userdata.StoreCode(code),
					Token: token,
				})
			}
		}

		data.CachedOnly = r.Form.Get("cached") == "on"

		isStoreStremThru := false
		for i := range data.Stores {
			if data.Stores[i].Code.IsStremThru() {
				isStoreStremThru = true
				break
			}
		}

		if !isStoreStremThru {
			for i := range data.Upstreams {
				up := &data.Upstreams[i]
				up.NoContentProxy = false
			}
		}
	}

	if IsPublicInstance && len(data.Upstreams) > MaxPublicInstanceUpstreamCount {
		data.Upstreams = data.Upstreams[0:MaxPublicInstanceUpstreamCount]
	}

	for i := range data.Upstreams {
		up := &data.Upstreams[i]
		if up.URL != "" {
			if baseUrl, err := stremio_addon.ExtractBaseURL(up.URL); err == nil {
				up.baseUrl = baseUrl
			}
		}
	}

	return data, nil
}
