package stremio_transformer

import "strings"

var StreamTemplateCinema = StreamTemplateBlob{
	Name: strings.TrimSpace(`{stream.resolution}`),
	Description: strings.TrimSpace(`{stream.title::exists["🎬 {stream.title}"||""]} {stream.year::exists["({stream.year})"||""]} {stream.seasonEpisode::exists["📗{stream.seasonEpisode::first}"||""]} {stream.seasonEpisode::exists["📘{stream.seasonEpisode::last}"||""]} 
{stream.quality::exists["🎥 {stream.quality} "||""]}{stream.encode::exists["🎞 {stream.encode} "||""]} {stream.audioTags::exists["🎧 {stream.audioTags::join(' | ')} "||""]}{stream.audioChannels::exists["🔊 {stream.audioChannels::join(' | ')}"||""]}
{stream.languages::exists["🌐 {stream.languages::join(' | ')} "||""]} 
{stream.size::>0["📦 {stream.size::bytes} "||""]}{stream.duration::>0["⏱ {stream.duration::time} "||""]}
{stream.seeders::>0["🌱 {stream.seeders} "||""]} {stream.message::exists["ℹ {stream.message}"||""]}
{addon.name} {service.cached::istrue["⚡"||""]}{service.cached::isfalse["❌"||""]}
`),
}.MustParse()
