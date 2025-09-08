package context

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/MunifTanjim/stremthru/store"
)

type storeContextKey struct{}

type StoreContext struct {
	Store             store.Store
	StoreAuthToken    string
	PeerToken         string
	IsProxyAuthorized bool
	ProxyAuthUser     string
	ProxyAuthPassword string
	ClientIP          string // optional

	Log *slog.Logger
}

func SetStoreContext(r *http.Request) *http.Request {
	ctx := context.WithValue(r.Context(), storeContextKey{}, &StoreContext{})
	return r.WithContext(ctx)
}

func GetStoreContext(r *http.Request) *StoreContext {
	return r.Context().Value(storeContextKey{}).(*StoreContext)
}
