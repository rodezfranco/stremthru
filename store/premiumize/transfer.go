package premiumize

import (
	"net/url"
	"path/filepath"
	"time"

	"github.com/rodezfranco/stremthru/internal/util"
)

type TranscodeStatus string

const (
	TranscodeStatusError         TranscodeStatus = "error"
	TranscodeStatusFetchPending  TranscodeStatus = "fetch_pending"
	TranscodeStatusFinished      TranscodeStatus = "finished"
	TranscodeStatusGoodAsIs      TranscodeStatus = "good_as_is"
	TranscodeStatusNotApplicable TranscodeStatus = "not_applicable"
	TranscodeStatusPending       TranscodeStatus = "pending"
	TranscodeStatusRunning       TranscodeStatus = "running"
)

type TransferStatus string

const (
	TransferStatusBanned   TransferStatus = "banned"
	TransferStatusDeleted  TransferStatus = "deleted"
	TransferStatusError    TransferStatus = "error"
	TransferStatusFinished TransferStatus = "finished"
	TransferStatusQueued   TransferStatus = "queued"
	TransferStatusRunning  TransferStatus = "running"
	TransferStatusSeeding  TransferStatus = "seeding"
	TransferStatusTimeout  TransferStatus = "timeout"
	TransferStatusWaiting  TransferStatus = "waiting"
)

type ListTransfersDataItem struct {
	Id       string         `json:"id"`
	Name     string         `json:"name"`
	Message  string         `json:"message"`
	Status   TransferStatus `json:"status"`
	Progress float32        `json:"progress"`
	Src      string         `json:"src"`
	FolderId string         `json:"folder_id"`
	FileId   string         `json:"file_id"`
}

func (t ListTransfersDataItem) GetAddedAt() time.Time {
	return time.Unix(0, 0).UTC()
}

type ListTransfersData struct {
	Transfers []ListTransfersDataItem `json:"transfers"`
}

type listTransfersData struct {
	ResponseContainer
	ListTransfersData
}

type ListTransfersParams struct {
	Ctx
}

func (c APIClient) ListTransfers(params *ListTransfersParams) (APIResponse[ListTransfersData], error) {
	response := &listTransfersData{}
	res, err := c.Request("GET", "/transfer/list", params, response)
	return newAPIResponse(res, response.ListTransfersData), err
}

type CreateTransferData struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type createTransferData struct {
	ResponseContainer
	CreateTransferData
}

type CreateTransferParams struct {
	Ctx
	Src      string
	FolderId string
}

func (c APIClient) CreateTransfer(params *CreateTransferParams) (APIResponse[CreateTransferData], error) {
	form := &url.Values{}
	form.Add("src", params.Src)
	if params.FolderId != "" {
		form.Add("folder_id", params.FolderId)

	}
	params.Form = form

	response := &createTransferData{}
	res, err := c.Request("POST", "/transfer/create", params, response)
	return newAPIResponse(res, response.CreateTransferData), err
}

type CreateDirectDownloadLinkDataContent struct {
	Path            string          `json:"path"`
	Size            int64           `json:"size"`
	Link            string          `json:"link"`
	StreamLink      string          `json:"stream_link"`
	TranscodeStatus TranscodeStatus `json:"transcode_status"`
}

func (f CreateDirectDownloadLinkDataContent) GetPath() string {
	path, _ := util.RemoveRootFolderFromPath(f.Path)
	return path
}

func (f CreateDirectDownloadLinkDataContent) GetName() string {
	return filepath.Base(f.Path)
}

type CreateDirectDownloadLinkData struct {
	Location string                                `json:"location"`
	Filename string                                `json:"filename"`
	Filesize int                                   `json:"filesize"`
	Content  []CreateDirectDownloadLinkDataContent `json:"content"`
}

type createDirectDownloadLinkData struct {
	ResponseContainer
	CreateDirectDownloadLinkData
}

type CreateDirectDownloadLinkParams struct {
	Ctx
	Src string
}

func (c APIClient) CreateDirectDownloadLink(params *CreateDirectDownloadLinkParams) (APIResponse[CreateDirectDownloadLinkData], error) {
	form := &url.Values{}
	form.Add("src", params.Src)
	params.Form = form

	response := &createDirectDownloadLinkData{}
	res, err := c.Request("POST", "/transfer/directdl", params, response)
	return newAPIResponse(res, response.CreateDirectDownloadLinkData), err
}

type DeleteTransferData struct{}

type deleteTransferData struct {
	ResponseContainer
	DeleteTransferData
}

type DeleteTransferParams struct {
	Ctx
	Id string
}

// also deletes the folder
func (c APIClient) DeleteTransfer(params *DeleteTransferParams) (APIResponse[DeleteTransferData], error) {
	form := &url.Values{}
	form.Add("id", params.Id)
	params.Form = form

	response := &deleteTransferData{}
	res, err := c.Request("POST", "/transfer/delete", params, response)
	return newAPIResponse(res, response.DeleteTransferData), err
}
