package server

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type reqCtxKey struct{}

type ReqCtx struct {
	StartTime time.Time
	RequestId string
	Error     error
	ReqPath   string
	ReqQuery  url.Values
	Log       *slog.Logger
}

func (ctx *ReqCtx) RedactURLPathValues(r *http.Request, names ...string) {
	for _, name := range names {
		if value := r.PathValue(name); value != "" {
			ctx.ReqPath = strings.Replace(ctx.ReqPath, value, "{"+name+"}", 1)
		}
	}
}

func (ctx *ReqCtx) RedactURLQueryParams(r *http.Request, names ...string) {
	for _, name := range names {
		if _, ok := ctx.ReqQuery[name]; ok {
			ctx.ReqQuery.Set(name, "...redacted...")
		}
	}
}

func SetReqCtx(r *http.Request, reqCtx *ReqCtx) *http.Request {
	ctx := context.WithValue(r.Context(), reqCtxKey{}, reqCtx)
	return r.WithContext(ctx)
}

func GetReqCtx(r *http.Request) *ReqCtx {
	return r.Context().Value(reqCtxKey{}).(*ReqCtx)
}
