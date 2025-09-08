package stremio_torz

import (
	"net/http"
	"strings"

	"github.com/rodezfranco/stremthru/internal/config"
	"github.com/rodezfranco/stremthru/internal/shared"
	stremio_shared "github.com/rodezfranco/stremthru/internal/stremio/shared"
	"github.com/rodezfranco/stremthru/stremio"
)

func GetManifest(r *http.Request, ud *UserData) *stremio.Manifest {
	isConfigured := ud.HasRequiredValues()

	id := shared.GetReversedHostname(r) + ".torz"
	name := "StremThru Torz"
	description := "Stremio Addon to access crowdsourced Torz"

	if isConfigured {
		storeHint := ""
		for i := range ud.Stores {
			code := string(ud.Stores[i].Code)
			if code == "" {
				code = "st"
			}
			if i > 0 {
				storeHint += " | "
			}
			storeHint += code
		}
		if storeHint != "" {
			storeHint = strings.ToUpper(storeHint)
		}

		description += " — " + storeHint
	}

	streamResource := stremio.Resource{
		Name: stremio.ResourceNameStream,
		Types: []stremio.ContentType{
			stremio.ContentTypeMovie,
			stremio.ContentTypeSeries,
		},
		IDPrefixes: []string{"tt"},
	}

	if config.Feature.IsEnabled(config.FeatureAnime) {
		streamResource.Types = append(streamResource.Types, "anime")
		streamResource.IDPrefixes = append(streamResource.IDPrefixes, "kitsu:", "mal:")
	}

	manifest := &stremio.Manifest{
		ID:          id,
		Name:        name,
		Description: description,
		Version:     config.Version,
		Resources: []stremio.Resource{
			streamResource,
		},
		Types:    []stremio.ContentType{},
		Catalogs: []stremio.Catalog{},
		Logo:     "https://emojiapi.dev/api/v1/sparkles/256.png",
		BehaviorHints: &stremio.BehaviorHints{
			Configurable:          true,
			ConfigurationRequired: !isConfigured,
		},
	}

	return manifest
}

func handleManifest(w http.ResponseWriter, r *http.Request) {
	if !IsMethod(r, http.MethodGet) {
		shared.ErrorMethodNotAllowed(r).Send(w, r)
		return
	}

	ud, err := getUserData(r)
	if err != nil {
		SendError(w, r, err)
		return
	}

	manifest := GetManifest(r, ud)

	stremio_shared.ClaimAddonOnStremioAddonsDotNet(manifest, "eyJhbGciOiJkaXIiLCJlbmMiOiJBMTI4Q0JDLUhTMjU2In0..TRRjkBGDOJQnN_AL7ngx3Q.mdQH2SF8ifXxQ3QsEkM1rCO7xuIxLDizZD9nla6nn-I4OJQw4ngjTUj98DhXXRM_L8F5frudd4Hwqrt3nS1b5DJnKU1wwAqDyw2ka7allZgLbKKNRznT2P_gkOICvjLD.vx5zJ9EsmaafrH5nlogAPQ")

	SendResponse(w, r, 200, manifest)
}
