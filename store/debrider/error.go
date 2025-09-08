package debrider

import (
	"github.com/rodezfranco/stremthru/core"
	"github.com/rodezfranco/stremthru/store"
)

func UpstreamErrorWithCause(cause error) *core.UpstreamError {
	err := core.NewUpstreamError("")
	err.StoreName = string(store.StoreNameDebrider)

	if rerr, ok := cause.(*ResponseContainer); ok {
		err.Msg = rerr.Message
		err.UpstreamCause = rerr
	} else {
		err.Cause = cause
	}

	return err
}
