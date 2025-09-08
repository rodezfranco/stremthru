package debrider

import (
	"path/filepath"

	"github.com/MunifTanjim/stremthru/internal/util"
)

type CheckLinkAvailabilityDataItemFile struct {
	Name         string `json:"name"`
	Size         int64  `json:"size"`
	DownloadLink string `json:"download_link"`
}

func (f *CheckLinkAvailabilityDataItemFile) GetPath() string {
	path, _ := util.RemoveRootFolderFromPath(f.Name)
	return path
}

func (f *CheckLinkAvailabilityDataItemFile) GetName() string {
	return filepath.Base(f.Name)
}

type CheckLinkAvailabilityDataItem struct {
	Cached bool                                `json:"cached"`
	Hash   string                              `json:"hash"`  // only when cached
	Files  []CheckLinkAvailabilityDataItemFile `json:"files"` // only when cached
}

type CheckLinkAvailabilityData struct {
	ResponseContainer
	Result []CheckLinkAvailabilityDataItem `json:"result"`
}

type CheckLinkAvailabilityParams struct {
	Ctx
	Data []string `json:"data"` // links
}

func (c APIClient) CheckLinkAvailability(params *CheckLinkAvailabilityParams) (APIResponse[CheckLinkAvailabilityData], error) {
	params.JSON = params
	response := &CheckLinkAvailabilityData{}
	res, err := c.Request("POST", "/v1/link/dlookup", params, response)
	return newAPIResponse(res, *response), err
}
