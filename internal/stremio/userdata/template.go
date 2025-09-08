package stremio_userdata

import "github.com/rodezfranco/stremthru/internal/stremio/configure"

type TemplateDataUserData struct {
	SavedUserDataKey     string
	SavedUserDataOptions []configure.ConfigOption
	IsRedacted           bool
}
