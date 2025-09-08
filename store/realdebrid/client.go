package realdebrid

import (
	"net/http"
	"net/url"

	"github.com/MunifTanjim/stremthru/core"
	"github.com/MunifTanjim/stremthru/internal/config"
	"github.com/MunifTanjim/stremthru/internal/request"
	"github.com/MunifTanjim/stremthru/store"
)

var DefaultHTTPClient = config.DefaultHTTPClient

type APIClientConfig struct {
	BaseURL    string // default: https://api.real-debrid.com
	APIKey     string
	HTTPClient *http.Client
	UserAgent  string
}

type APIClient struct {
	BaseURL    *url.URL
	HTTPClient *http.Client
	apiKey     string
	agent      string
}

func NewAPIClient(conf *APIClientConfig) *APIClient {
	if conf.UserAgent == "" {
		conf.UserAgent = "stremthru"
	}

	if conf.BaseURL == "" {
		conf.BaseURL = "https://api.real-debrid.com"
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

	return c
}

type Ctx = request.Ctx

func (c APIClient) newRequest(method, path string, params request.Context) (req *http.Request, err error) {
	if params == nil {
		params = &Ctx{}
	}

	url := c.BaseURL.JoinPath(path)

	query := url.Query()

	body, contentType, err := params.PrepareBody(method, &query)
	if err != nil {
		return nil, err
	}

	url.RawQuery = query.Encode()

	req, err = http.NewRequestWithContext(params.GetContext(), method, url.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+params.GetAPIKey(c.apiKey))
	req.Header.Add("User-Agent", c.agent)
	if len(contentType) > 0 {
		req.Header.Add("Content-Type", contentType)
	}

	return req, nil
}

func (c APIClient) Request(method, path string, params request.Context, v ResponseContainer) (*http.Response, error) {
	req, err := c.newRequest(method, path, params)
	if err != nil {
		error := core.NewStoreError("failed to create request")
		error.StoreName = string(store.StoreNameRealDebrid)
		error.Cause = err
		return nil, error
	}
	res, err := c.HTTPClient.Do(req)
	err = processResponseBody(res, err, v)
	if err != nil {
		err := UpstreamErrorWithCause(err)
		err.InjectReq(req)
		if res != nil {
			err.StatusCode = res.StatusCode
		}
		return res, err
	}
	return res, nil
}
