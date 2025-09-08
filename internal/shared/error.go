package shared

import (
	"net/http"

	"github.com/rodezfranco/stremthru/core"
)

var ErrorUnauthorized = func(r *http.Request) *core.APIError {
	err := core.NewAPIError("unauthorized")
	err.InjectReq(r)
	err.Code = core.ErrorCodeUnauthorized
	err.StatusCode = http.StatusUnauthorized
	return err
}

var ErrorForbidden = func(r *http.Request) *core.APIError {
	err := core.NewAPIError("forbidden")
	err.InjectReq(r)
	err.Code = core.ErrorCodeForbidden
	err.StatusCode = http.StatusForbidden
	return err
}

var ErrorNotFound = func(r *http.Request) *core.APIError {
	err := core.NewAPIError("not found")
	err.InjectReq(r)
	err.Code = core.ErrorCodeNotFound
	err.StatusCode = http.StatusNotFound
	return err
}

var ErrorMethodNotAllowed = func(r *http.Request) *core.APIError {
	err := core.NewAPIError("method not allowed")
	err.InjectReq(r)
	err.Code = core.ErrorCodeMethodNotAllowed
	err.StatusCode = http.StatusMethodNotAllowed
	return err
}

var ErrorUnsupportedMediaType = func(r *http.Request) *core.APIError {
	err := core.NewAPIError("unsupported media type")
	err.InjectReq(r)
	err.Code = core.ErrorCodeUnsupportedMediaType
	err.StatusCode = http.StatusUnsupportedMediaType
	return err
}

var ErrorProxyAuthRequired = func(r *http.Request) *core.APIError {
	err := core.NewAPIError("proxy auth required")
	err.InjectReq(r)
	err.Code = core.ErrorCodeProxyAuthenticationRequired
	err.StatusCode = http.StatusProxyAuthRequired
	return err
}

var ErrorBadRequest = func(r *http.Request, msg string) *core.APIError {
	if msg == "" {
		msg = "bad request"
	}

	err := core.NewAPIError(msg)
	err.InjectReq(r)
	err.Code = core.ErrorCodeBadRequest
	err.StatusCode = http.StatusBadRequest
	return err
}

var ErrorInternalServerError = func(r *http.Request, msg string) *core.APIError {
	if msg == "" {
		msg = "internal server error"
	}

	err := core.NewAPIError(msg)
	err.InjectReq(r)
	err.Code = core.ErrorCodeInternalServerError
	err.StatusCode = http.StatusInternalServerError
	return err
}

var ErrorBadGateway = func(r *http.Request, msg string) *core.APIError {
	if msg == "" {
		msg = "bad gateway"
	}

	err := core.NewAPIError(msg)
	err.InjectReq(r)
	err.Code = core.ErrorCodeBadGateway
	err.StatusCode = http.StatusBadGateway
	return err
}
