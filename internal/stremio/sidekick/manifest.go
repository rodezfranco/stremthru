package stremio_sidekick

import (
	"net/http"

	"github.com/rodezfranco/stremthru/internal/config"
	"github.com/rodezfranco/stremthru/internal/shared"
	"github.com/rodezfranco/stremthru/stremio"
)

func GetManifest(r *http.Request) *stremio.Manifest {
	return &stremio.Manifest{
		ID:          shared.GetReversedHostname(r) + ".sidekick",
		Name:        "Stremio Sidekick",
		Description: "Extra Features for Stremio",
		Version:     config.Version,
		Logo:        "https://emojiapi.dev/api/v1/sparkles/256.png",
		Resources:   []stremio.Resource{},
		Types:       []stremio.ContentType{},
		Catalogs:    []stremio.Catalog{},
		BehaviorHints: &stremio.BehaviorHints{
			Configurable:          true,
			ConfigurationRequired: true,
		},
	}
}
