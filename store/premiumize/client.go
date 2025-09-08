package premiumize

import (
	"net/http"
	"net/url"

	"github.com/rodezfranco/stremthru/core"
	"github.com/rodezfranco/stremthru/internal/config"
	"github.com/rodezfranco/stremthru/internal/request"
	"github.com/rodezfranco/stremthru/store"
)

var DefaultHTTPClient = config.DefaultHTTPClient

type APIClientConfig struct {
	BaseURL    string // default: https://www.premiumize.me/api
	APIKey     string
	HTTPClient *http.Client
	UserAgent  string
}

type APIClient struct {
	BaseURL    *url.URL
	HTTPClient *http.Client
	apiKey     string
	agent      string
	reqQuery   func(query *url.Values, params request.Context)
	reqHeader  func(query *http.Header, params request.Context)
}

func NewAPIClient(conf *APIClientConfig) *APIClient {
	if conf.UserAgent == "" {
		conf.UserAgent = "stremthru"
	}

	if conf.BaseURL == "" {
		conf.BaseURL = "https://www.premiumize.me/api"
	}

	if conf.HTTPClient == nil {
		conf.HTTPClient = DefaultHTTPClient
	}

	c := &APIClient{}

	baseUrl, err := url.Parse(conf.BaseURL)
	if err != nil {
		panic(err)
	}

	c.BaseURL = baseUrl
	c.HTTPClient = conf.HTTPClient
	c.apiKey = conf.APIKey
	c.agent = conf.UserAgent

	c.reqQuery = func(query *url.Values, params request.Context) {
		query.Add("apikey", params.GetAPIKey(c.apiKey))
	}

	c.reqHeader = func(header *http.Header, params request.Context) {
		header.Add("User-Agent", c.agent)
	}

	return c
}

type Ctx = request.Ctx

func (c APIClient) Request(method, path string, params request.Context, v ResponseEnvelop) (*http.Response, error) {
	if params == nil {
		params = &Ctx{}
	}
	req, err := params.NewRequest(c.BaseURL, method, path, c.reqHeader, c.reqQuery)
	if err != nil {
		error := core.NewStoreError("failed to create request")
		error.StoreName = string(store.StoreNamePremiumize)
		error.Cause = err
		return nil, error
	}
	res, err := c.HTTPClient.Do(req)
	err = processResponseBody(res, err, v)
	if err != nil {
		err := UpstreamErrorWithCause(err)
		err.InjectReq(req)
		if res != nil && res.StatusCode >= http.StatusBadRequest {
			err.StatusCode = res.StatusCode
		}
		if err.StatusCode < http.StatusBadRequest {
			err.StatusCode = http.StatusBadRequest
		}
		return res, err
	}
	return res, nil
}
