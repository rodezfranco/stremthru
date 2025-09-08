package stremio_store

import (
	"net/http"

	"github.com/rodezfranco/stremthru/core"
	"github.com/rodezfranco/stremthru/internal/logger"
	"github.com/rodezfranco/stremthru/internal/server"
)

var log = logger.Scoped("stremio/store")

var pttLog = logger.Scoped("stremio/store/ptt")

func LogError(r *http.Request, msg string, err error) {
	ctx := server.GetReqCtx(r)
	ctx.Log.Error(msg, "error", core.PackError(err))
}
