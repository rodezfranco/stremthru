package stremio_transformer

import "strings"

var StreamTemplateCinema = StreamTemplateBlob{
	Name: strings.TrimSpace(`{{.Resolution}}`),
	Description: strings.TrimSpace(`
{{if ne .Title ""}}🎬 {{.Title}}{{end}} {{if ne .Year ""}}({{.Year}}){{end}}
{{if ne (index .SeasonEpisode 0) ""}}📗 {{index .SeasonEpisode 0}}{{end}} {{if ne (index .SeasonEpisode 1) ""}}📘 {{index .SeasonEpisode 1}}{{end}}
{{if ne .Quality ""}}🎥 {{.Quality}}{{end}} {{if ne .Encode ""}}🎞 {{.Encode}}{{end}}
{{if gt (len .AudioTags) 0}}🎧 {{str_join .AudioTags " | "}}{{end}} {{if gt (len .AudioChannels) 0}}🔊 {{str_join .AudioChannels " | "}}{{end}}
{{if gt (len .Languages) 0}}🌐 {{str_join .Languages " | "}}{{end}}
{{if ne .Size ""}}📦 {{.Size}}{{end}} {{if ne .Duration ""}}⏱ {{.Duration}}{{end}}
{{if ne .Seeders ""}}🌱 {{.Seeders}}{{end}} {{if ne .Message ""}}ℹ {{.Message}}{{end}}
{{.AddonName}} {{if .Service.Cached}}⚡{{else}}❌{{end}}
`),
}.MustParse()
