package oauth

import (
	"github.com/rodezfranco/stremthru/internal/logger"
)

var log = logger.Scoped("oauth")
var traktLog = logger.Scoped("oauth/trakt")
var kitsuLog = logger.Scoped("oauth/kitsu")
var tokenSourceLog = logger.Scoped("oauth/token_source")
