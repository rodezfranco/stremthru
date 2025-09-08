package offcloud

import (
	"github.com/rodezfranco/stremthru/core"
	"github.com/rodezfranco/stremthru/store"
)

func UpstreamErrorWithCause(cause error) *core.UpstreamError {
	err := core.NewUpstreamError("")
	err.StoreName = string(store.StoreNameOffcloud)

	if rerr, ok := cause.(*ResponseContainer); ok {
		err.Msg = rerr.Err
		err.UpstreamCause = rerr
	} else {
		err.Cause = cause
	}

	return err
}
