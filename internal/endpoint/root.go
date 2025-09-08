package endpoint

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/rodezfranco/stremthru/internal/config"
	stremio_shared "github.com/rodezfranco/stremthru/internal/stremio/shared"
)

//go:embed root.html
var templateBlob string

type rootTemplateDataAddon struct {
	Name string
	URL  string
}

type rootTemplateDataSection struct {
	Title   string        `json:"title"`
	Content template.HTML `json:"content"`
}

type RootTemplateData struct {
	Title       string                    `json:"-"`
	Description template.HTML             `json:"description"`
	Version     string                    `json:"-"`
	Addons      []rootTemplateDataAddon   `json:"-"`
	Sections    []rootTemplateDataSection `json:"sections"`
}

var rootTemplateData = func() RootTemplateData {
	td := RootTemplateData{}
	err := json.Unmarshal([]byte(config.LandingPage), &td)
	if err != nil {
		panic("malformed config for landing page: " + config.LandingPage)
	}
	return td
}()

var ExecuteTemplate = func() func(data *RootTemplateData) (bytes.Buffer, error) {
	tmpl := template.Must(template.New("root.html").Parse(templateBlob))
	return func(data *RootTemplateData) (bytes.Buffer, error) {
		var buf bytes.Buffer
		err := tmpl.Execute(&buf, data)
		return buf, err
	}
}()

func handleRoot(w http.ResponseWriter, r *http.Request) {
	td := &RootTemplateData{
		Title:       "StremThru",
		Description: rootTemplateData.Description,
		Version:     config.Version,
		Addons:      []rootTemplateDataAddon{},
		Sections:    rootTemplateData.Sections,
	}
	addons := stremio_shared.GetStremThruAddons()
	for _, addon := range addons {
		td.Addons = append(td.Addons, rootTemplateDataAddon{
			Name: addon.Name,
			URL:  addon.URL,
		})
	}

	buf, err := ExecuteTemplate(td)
	if err != nil {
		SendError(w, r, err)
		return
	}
	SendHTML(w, 200, buf)
}

func AddRootEndpoint(mux *http.ServeMux) {
	mux.HandleFunc("/{$}", handleRoot)
}
