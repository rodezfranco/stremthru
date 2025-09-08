package easydebrid

import (
	"net/http"
	"path/filepath"

	"github.com/rodezfranco/stremthru/internal/util"
)

type LookupLinkDetailsDataItemFile struct {
	Size   int64  `json:"size"`
	Name   string `json:"name"`
	Folder string `json:"folder"`
}

func (f *LookupLinkDetailsDataItemFile) GetPath() string {
	path, _ := util.RemoveRootFolderFromPath(filepath.Join(f.Folder, f.Name))
	return path
}

func (f *LookupLinkDetailsDataItemFile) GetName() string {
	return f.Name
}

type LookupLinkDetailsDataItem struct {
	ResponseContainer
	Cached bool                            `json:"cached"`
	Files  []LookupLinkDetailsDataItemFile `json:"files"`
}

func (t LookupLinkDetailsDataItem) GetName() string {
	if len(t.Files) > 0 {
		f := &t.Files[0]
		if _, rootFolder := util.RemoveRootFolderFromPath(filepath.Join(f.Folder, f.Name)); rootFolder != "" {
			return rootFolder
		}
	}
	return ""
}

type LookupLinkDetailsData struct {
	ResponseContainer
	Result []LookupLinkDetailsDataItem `json:"result"`
}

type LookupLinkDetailsParams struct {
	Ctx
	URLs []string `json:"urls"`
}

func (c APIClient) LookupLinkDetails(params *LookupLinkDetailsParams) (APIResponse[LookupLinkDetailsData], error) {
	params.JSON = params
	response := &LookupLinkDetailsData{}
	res, err := c.Request("POST", "/v1/link/lookupdetails", params, response)
	return newAPIResponse(res, *response), err
}

type GenerateLinkDataFile struct {
	Filename  string   `json:"filename"`
	Directory []string `json:"directory"`
	Size      int64    `json:"size"`
	URL       string   `json:"url"`
}

type GenerateLinkData struct {
	ResponseContainer
	Files []GenerateLinkDataFile `json:"files"`
}

type GenerateLinkParams struct {
	Ctx
	URL      string `json:"url"`
	ClientIP string `json:"-"`
}

func (c APIClient) GenerateLink(params *GenerateLinkParams) (APIResponse[GenerateLinkData], error) {
	params.JSON = params
	if params.ClientIP != "" {
		if params.Headers == nil {
			params.Headers = &http.Header{}
		}
		params.Headers.Add("X-Forwarded-For", params.ClientIP)
	}
	response := &GenerateLinkData{}
	res, err := c.Request("POST", "/v1/link/generate", params, response)
	return newAPIResponse(res, *response), err

}
