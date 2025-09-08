package endpoint

import (
	"net/http"

	"github.com/rodezfranco/stremthru/internal/config"
	"github.com/rodezfranco/stremthru/internal/stremio/disabled"
	stremio_list "github.com/rodezfranco/stremthru/internal/stremio/list"
	"github.com/rodezfranco/stremthru/internal/stremio/root"
	"github.com/rodezfranco/stremthru/internal/stremio/sidekick"
	"github.com/rodezfranco/stremthru/internal/stremio/store"
	stremio_torz "github.com/rodezfranco/stremthru/internal/stremio/torz"
	"github.com/rodezfranco/stremthru/internal/stremio/wrap"
)

func AddStremioEndpoints(mux *http.ServeMux) {
	stremio_root.AddStremioEndpoints(mux)

	if config.Feature.IsEnabled(config.FeatureStremioList) {
		stremio_list.AddEndpoints(mux)
	}
	if config.Feature.IsEnabled(config.FeatureStremioStore) {
		stremio_store.AddStremioStoreEndpoints(mux)
	}
	if config.Feature.IsEnabled(config.FeatureStremioWrap) {
		stremio_wrap.AddStremioWrapEndpoints(mux)
	}
	if config.Feature.IsEnabled(config.FeatureStremioSidekick) {
		stremio_sidekick.AddStremioSidekickEndpoints(mux)
		stremio_disabled.AddStremioDisabledEndpoints(mux)
	}
	if config.Feature.IsEnabled(config.FeatureStremioTorz) {
		stremio_torz.AddStremioTorzEndpoints(mux)
	}
}
