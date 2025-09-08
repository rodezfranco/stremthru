package endpoint

import (
	"net/http"

	meta_id_map "github.com/MunifTanjim/stremthru/internal/meta/id_map"
	meta_letterboxd "github.com/MunifTanjim/stremthru/internal/meta/letterboxd"
)

func AddMetaEndpoints(mux *http.ServeMux) {
	meta_id_map.AddEndpoints(mux)
	meta_letterboxd.AddEndpoints(mux)
}
