package endpoint

import (
	"net/http"
	"strings"

	"github.com/rodezfranco/stremthru/core"
	"github.com/rodezfranco/stremthru/internal/config"
	"github.com/rodezfranco/stremthru/internal/context"
	"github.com/rodezfranco/stremthru/internal/server"
	"github.com/rodezfranco/stremthru/internal/shared"
	"github.com/rodezfranco/stremthru/store"
)

func withStoreContext(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, context.SetStoreContext(r))
	})
}

func StoreMiddleware(middlewares ...shared.MiddlewareFunc) shared.MiddlewareFunc {
	return shared.Middleware(append([]shared.MiddlewareFunc{withStoreContext}, middlewares...)...)
}

func extractProxyAuthToken(r *http.Request, readQuery bool) (token string, hasToken bool) {
	token = r.Header.Get(server.HEADER_STREMTHRU_AUTHORIZATION)
	if token == "" {
		token = r.Header.Get(server.HEADER_PROXY_AUTHORIZATION)
		if token != "" {
			r.Header.Del(server.HEADER_PROXY_AUTHORIZATION)
		}
	}
	if token == "" && readQuery {
		token = r.URL.Query().Get("token")
	}
	token = strings.TrimPrefix(token, "Basic ")
	return token, token != ""
}

func getProxyAuthorization(r *http.Request, readQuery bool) (isAuthorized bool, user, pass string) {
	token, hasToken := extractProxyAuthToken(r, readQuery)
	auth, err := core.ParseBasicAuth(token)
	isAuthorized = hasToken && err == nil && config.ProxyAuthPassword.GetPassword(auth.Username) == auth.Password
	user = auth.Username
	pass = auth.Password
	return isAuthorized, user, pass
}

func ProxyAuthContext(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.GetStoreContext(r)
		ctx.IsProxyAuthorized, ctx.ProxyAuthUser, ctx.ProxyAuthPassword = getProxyAuthorization(r, false)
		next.ServeHTTP(w, r)
	})
}

func getStoreName(r *http.Request) (store.StoreName, *core.StoreError) {
	name := r.Header.Get("X-StremThru-Store-Name")
	if name == "" {
		ctx := context.GetStoreContext(r)
		if ctx.IsProxyAuthorized {
			name = config.StoreAuthToken.GetPreferredStore(ctx.ProxyAuthUser)
			r.Header.Set("X-StremThru-Store-Name", name)
		}
	}
	if name == "" {
		return "", nil
	}
	return store.StoreName(name).Validate()
}

func getStoreAuthToken(r *http.Request) string {
	authHeader := r.Header.Get("X-StremThru-Store-Authorization")
	if authHeader == "" {
		authHeader = r.Header.Get("Authorization")
	}
	if authHeader == "" {
		ctx := context.GetStoreContext(r)
		if ctx.IsProxyAuthorized && ctx.Store != nil {
			if token := config.StoreAuthToken.GetToken(ctx.ProxyAuthUser, string(ctx.Store.GetName())); token != "" {
				return token
			}
		}
	}
	_, token, _ := strings.Cut(authHeader, " ")
	return strings.TrimSpace(token)
}

func getStore(r *http.Request) (store.Store, error) {
	name, err := getStoreName(r)
	if err != nil {
		err.InjectReq(r)
		err.StatusCode = http.StatusBadRequest
		return nil, err
	}
	return shared.GetStore(string(name)), nil
}

func StoreContext(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		store, err := getStore(r)
		if err != nil {
			SendError(w, r, err)
			return
		}
		ctx := context.GetStoreContext(r)
		ctx.Store = store
		ctx.StoreAuthToken = getStoreAuthToken(r)
		ctx.PeerToken = r.Header.Get("X-StremThru-Peer-Token")

		ctx.ClientIP = shared.GetClientIP(r, ctx)

		w.Header().Add("X-StremThru-Store-Name", r.Header.Get("X-StremThru-Store-Name"))
		next.ServeHTTP(w, r)
	})
}

func StoreRequired(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.GetStoreContext(r)

		if ctx.Store == nil {
			shared.ErrorBadRequest(r, "missing store").Send(w, r)
			return
		}

		if ctx.StoreAuthToken == "" {
			w.Header().Add("WWW-Authenticate", "Bearer realm=\"store:"+string(ctx.Store.GetName())+"\"")
			shared.ErrorUnauthorized(r).Send(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func AdminAuthed(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Basic "))
		if token == "" {
			shared.ErrorUnauthorized(r).Send(w, r)
			return
		}
		if auth, err := core.ParseBasicAuth(token); err != nil || config.AdminPassword.GetPassword(auth.Username) != auth.Password {
			shared.ErrorUnauthorized(r).Send(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
