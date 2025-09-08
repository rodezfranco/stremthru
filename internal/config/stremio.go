package config

import (
	"strings"

	"github.com/MunifTanjim/stremthru/internal/util"
)

type stremioConfigList struct {
	PublicMaxListCount int
}

type stremioConfigTorz struct {
	LazyPull            bool
	PublicMaxStoreCount int
}

type stremioConfigWrap struct {
	PublicMaxUpstreamCount int
	PublicMaxStoreCount    int
}

type StremioConfig struct {
	List stremioConfigList
	Torz stremioConfigTorz
	Wrap stremioConfigWrap
}

func parseStremio() StremioConfig {
	stremio := StremioConfig{
		List: stremioConfigList{
			PublicMaxListCount: util.MustParseInt(getEnv("STREMTHRU_STREMIO_LIST_PUBLIC_MAX_LIST_COUNT")),
		},
		Torz: stremioConfigTorz{
			LazyPull:            strings.ToLower(getEnv("STREMTHRU_STREMIO_TORZ_LAZY_PULL")) == "true",
			PublicMaxStoreCount: util.MustParseInt(getEnv("STREMTHRU_STREMIO_TORZ_PUBLIC_MAX_STORE_COUNT")),
		},
		Wrap: stremioConfigWrap{
			PublicMaxUpstreamCount: util.MustParseInt(getEnv("STREMTHRU_STREMIO_WRAP_PUBLIC_MAX_UPSTREAM_COUNT")),
			PublicMaxStoreCount:    util.MustParseInt(getEnv("STREMTHRU_STREMIO_WRAP_PUBLIC_MAX_STORE_COUNT")),
		},
	}
	return stremio
}

var Stremio = parseStremio()
