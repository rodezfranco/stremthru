package stremio_list

import (
	"github.com/rodezfranco/stremthru/internal/shared"
	stremio_shared "github.com/rodezfranco/stremthru/internal/stremio/shared"
)

var IsMethod = shared.IsMethod
var SendError = shared.SendError
var ExtractRequestBaseURL = shared.ExtractRequestBaseURL

var SendResponse = stremio_shared.SendResponse
var SendHTML = stremio_shared.SendHTML
var GetPathValue = stremio_shared.GetPathValue
