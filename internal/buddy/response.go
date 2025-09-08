package buddy

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/rodezfranco/stremthru/core"
)

type ResponseError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	StatusCode int       `json:"status_code"`
}

func (e *ResponseError) Error() string {
	ret, _ := json.Marshal(e)
	return string(ret)
}

type Response[T any] struct {
	Data  T              `json:"data,omitempty"`
	Error *ResponseError `json:"error,omitempty"`
}

type ResponseEnvelop interface {
	GetError() error
}

func (r Response[any]) GetError() error {
	if r.Error == nil {
		return nil
	}
	return r.Error
}

type APIResponse[T any] struct {
	Header     http.Header
	StatusCode int
	Data       T
}

func newAPIResponse[T any](res *http.Response, data T) APIResponse[T] {
	apiResponse := APIResponse[T]{
		StatusCode: 503,
		Data:       data,
	}
	if res != nil {
		apiResponse.Header = res.Header
		apiResponse.StatusCode = res.StatusCode
	}
	return apiResponse
}

func processResponseBody(res *http.Response, err error, v ResponseEnvelop) error {
	if err != nil {
		return err
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()

	if err != nil {
		return err
	}

	err = core.UnmarshalJSON(res.StatusCode, body, v)
	if err != nil {
		return err
	}

	return v.GetError()
}
