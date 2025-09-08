package stremio_list

import (
	"github.com/rodezfranco/stremthru/internal/mdblist"
)

var mdblistClient = mdblist.NewAPIClient(&mdblist.APIClientConfig{})
