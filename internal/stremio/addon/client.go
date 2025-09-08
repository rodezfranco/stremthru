package stremio_addon

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rodezfranco/stremthru/core"
	"github.com/rodezfranco/stremthru/internal/config"
	"github.com/rodezfranco/stremthru/internal/request"
	"github.com/rodezfranco/stremthru/internal/shared"
	"github.com/rodezfranco/stremthru/stremio"
	"golang.org/x/sync/singleflight"
)

var DefaultHTTPClient = func() *http.Client {
	transport := config.DefaultHTTPTransport.Clone()
	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
}()

type ClientConfig struct {
	HTTPClient *http.Client
}

type Client struct {
	HTTPClient *http.Client

	reqQuery  func(query *url.Values, params request.Context)
	reqHeader func(query *http.Header, params request.Context)
}

var commonHeaders = func() map[string]string {
	encodedHeaders := map[string]string{
		"UmVmZXJlcg==":     "aHR0cHM6Ly93ZWIuc3RyZW1pby5jb20v",
		"VXNlci1BZ2VudA==": "TW96aWxsYS81LjAgKE1hY2ludG9zaDsgSW50ZWwgTWFjIE9TIFggMTBfMTVfNykgQXBwbGVXZWJLaXQvNTM3LjM2IChLSFRNTCwgbGlrZSBHZWNrbykgQ2hyb21lLzEzNC4wLjAuMCBTYWZhcmkvNTM3LjM2",
	}
	headers := map[string]string{}
	for k, v := range encodedHeaders {
		key, err := core.Base64Decode(k)
		if err != nil {
			panic(err)
		}
		val, err := core.Base64Decode(v)
		if err != nil {
			panic(err)
		}
		headers[key] = val
	}
	return headers
}()

func NewClient(conf *ClientConfig) *Client {
	if conf.HTTPClient == nil {
		conf.HTTPClient = DefaultHTTPClient
	}

	c := &Client{}

	c.HTTPClient = conf.HTTPClient

	c.reqQuery = func(query *url.Values, params request.Context) {
	}

	c.reqHeader = func(header *http.Header, params request.Context) {
		for k, v := range commonHeaders {
			header.Set(k, v)
		}
	}

	return c
}

type Ctx = request.Ctx

type ResponseError struct {
	Body       string `json:"body"`
	StatusCode int    `json:"status_code"`
}

func (e *ResponseError) Error() string {
	ret, _ := json.Marshal(e)
	return string(ret)
}

func processResponseBody(res *http.Response, err error, v any) error {
	if err != nil {
		return err
	}

	contentType := res.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		err := core.NewAPIError("unxpected content-type: " + contentType)
		err.StatusCode = res.StatusCode
		return err
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()

	if err != nil {
		return err
	}

	if res.StatusCode >= 400 {
		return &ResponseError{
			Body:       string(body),
			StatusCode: res.StatusCode,
		}
	}

	return core.UnmarshalJSON(res.StatusCode, body, v)
}

func (c Client) Request(method string, url *url.URL, params request.Context, v any) (*http.Response, error) {
	if params == nil {
		params = &Ctx{}
	}
	req, err := params.NewRequest(url, method, "", c.reqHeader, c.reqQuery)
	if err != nil {
		error := core.NewAPIError("failed to create request")
		error.Cause = err
		return nil, error
	}
	res, err := c.HTTPClient.Do(req)
	err = processResponseBody(res, err, v)
	if err != nil {
		error := core.NewUpstreamError("")
		if rerr, ok := err.(*core.Error); ok {
			error.Msg = rerr.Msg
			error.Code = rerr.Code
			error.StatusCode = rerr.StatusCode
			error.UpstreamCause = rerr
		} else {
			error.Cause = err
		}
		error.InjectReq(req)
		return res, err
	}
	return res, nil
}

func adjustClientIPHeader(params request.Ctx, clientIp string, r *http.Request) {
	if clientIp == "" {
		if r != nil {
			r.Header.Del("X-Client-Ip")
			r.Header.Del("X-Forwarded-For")
			r.Header.Del("X-Real-IP")
		}
		return
	}

	if params.Headers == nil {
		params.Headers = &http.Header{}
	}

	if clientIp == "" {
		params.Headers.Del("X-Client-Ip")
		params.Headers.Del("X-Forwarded-For")
		params.Headers.Del("X-Real-IP")
	} else {
		params.Headers.Set("X-Client-Ip", clientIp)
		params.Headers.Set("X-Forwarded-For", clientIp)
		params.Headers.Set("X-Real-IP", clientIp)
	}
}

type GetManifestParams struct {
	request.Ctx
	BaseURL  *url.URL
	ClientIP string
}

func (c Client) GetManifest(params *GetManifestParams) (request.APIResponse[stremio.Manifest], error) {
	adjustClientIPHeader(params.Ctx, params.ClientIP, nil)
	response := &stremio.Manifest{}
	res, err := c.Request("GET", params.BaseURL.JoinPath("manifest.json"), params, response)
	if err == nil && !response.IsValid() {
		err = errors.New("invalid manifest")
	}
	return request.NewAPIResponse(res, *response), err
}

type FetchStreamParams struct {
	request.Ctx
	BaseURL  *url.URL
	Type     string
	Id       string
	ClientIP string
}

var fetchStreamGroup singleflight.Group

func (c Client) FetchStream(params *FetchStreamParams) (request.APIResponse[stremio.StreamHandlerResponse], error) {
	path := "stream/" + params.Type + "/" + params.Id
	url := params.BaseURL.JoinPath(path)
	apiResponse, err, _ := fetchStreamGroup.Do(url.String(), func() (any, error) {
		adjustClientIPHeader(params.Ctx, params.ClientIP, nil)
		response := &stremio.StreamHandlerResponse{}
		res, err := c.Request("GET", url, params, response)
		return request.NewAPIResponse(res, *response), err
	})
	return apiResponse.(request.APIResponse[stremio.StreamHandlerResponse]), err
}

type FetchCatalogParams struct {
	request.Ctx
	BaseURL  *url.URL
	Type     string
	Id       string
	Extra    string
	ClientIP string
}

func (c Client) FetchCatalog(params *FetchCatalogParams) (request.APIResponse[stremio.CatalogHandlerResponse], error) {
	path := "catalog/" + params.Type + "/" + params.Id
	if params.Extra != "" {
		path = path + "/" + params.Extra
	}
	adjustClientIPHeader(params.Ctx, params.ClientIP, nil)
	response := &stremio.CatalogHandlerResponse{}
	res, err := c.Request("GET", params.BaseURL.JoinPath(path), params, response)
	return request.NewAPIResponse(res, *response), err
}

type FetchMetaParams struct {
	request.Ctx
	BaseURL  *url.URL
	Type     string
	Id       string
	ClientIP string
}

func (c Client) FetchMeta(params *FetchMetaParams) (request.APIResponse[stremio.MetaHandlerResponse], error) {
	path := "meta/" + params.Type + "/" + params.Id
	adjustClientIPHeader(params.Ctx, params.ClientIP, nil)
	response := &stremio.MetaHandlerResponse{}
	res, err := c.Request("GET", params.BaseURL.JoinPath(path), params, response)
	return request.NewAPIResponse(res, *response), err
}

type FetchSubtitlesParams struct {
	request.Ctx
	BaseURL  *url.URL
	Type     string
	Id       string
	Extra    string
	ClientIP string
}

func (c Client) FetchSubtitles(params *FetchSubtitlesParams) (request.APIResponse[stremio.SubtitlesHandlerResponse], error) {
	path := "subtitles/" + params.Type + "/" + params.Id
	if params.Extra != "" {
		path = path + "/" + params.Extra
	}
	adjustClientIPHeader(params.Ctx, params.ClientIP, nil)
	response := &stremio.SubtitlesHandlerResponse{}
	res, err := c.Request("GET", params.BaseURL.JoinPath(path), params, response)
	return request.NewAPIResponse(res, *response), err
}

type ProxyResourceParams struct {
	request.Ctx
	BaseURL  *url.URL
	Resource string
	Type     string
	Id       string
	Extra    string
	ClientIP string
}

func (c Client) ProxyResource(w http.ResponseWriter, r *http.Request, params *ProxyResourceParams) {
	path := params.Resource + "/" + params.Type + "/" + params.Id
	if params.Extra != "" {
		path = path + "/" + params.Extra
	}
	adjustClientIPHeader(params.Ctx, params.ClientIP, r)
	w.Header().Del("Access-Control-Allow-Origin")
	shared.ProxyResponse(w, r, params.BaseURL.JoinPath(path).String(), config.TUNNEL_TYPE_AUTO)
}

func NormalizeManifestURL(manifestUrl string) (string, error) {
	u, err := url.Parse(manifestUrl)
	if err != nil {
		return manifestUrl, err
	}
	if u.Scheme == "stremio" {
		u.Scheme = "https"
	}
	if strings.HasSuffix(u.Path, "/configure") {
		u.Path = strings.TrimSuffix(u.Path, "/configure") + "/manifest.json"
	}
	return u.String(), nil
}

func ExtractBaseURL(manifestUrl string) (*url.URL, error) {
	u, err := url.Parse(manifestUrl)
	if err != nil {
		return nil, err
	}
	if !strings.HasSuffix(u.Path, "/manifest.json") {
		return nil, errors.New("invalid manifest url")
	}
	if u.RawPath != "" && u.RawPath != u.Path {
		u.RawPath = strings.TrimSuffix(u.RawPath, "/manifest.json")
		if u.Path, err = url.PathUnescape(u.RawPath); err != nil {
			return nil, err
		}
	} else {
		u.Path = strings.TrimSuffix(u.Path, "/manifest.json")
	}
	return u, nil
}
