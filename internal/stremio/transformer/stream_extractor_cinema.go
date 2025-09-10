package stremio_transformer

import "strings"

var StreamExtractorCinema = StreamExtractorBlob(strings.TrimSpace(`
name
(?i)^(?<resolution>.+)$

description
^(?:🎬 (?<t_title>.+?))?(?: \((?<year>\d{4})\))?(?: 📗(?<season>\d+))?(?: 📘(?<episode>\d+))?\s*
(?:🎥 (?<quality>.+?))?(?: 🎞 (?<encode>.+?))?\s*
(?:🎧 (?<audioTags>[^🔊\n]+))?(?: 🔊 (?<audioChannels>.+?))?\s*
(?:🌐 (?<languages>.+?))?\s*
(?:📦 (?<size>.+?))?(?: ⏱ (?<duration>.+?))?\s*
(?:🌱 (?<seeders>\d+))?(?: ℹ (?<message>.+?))?\s*
(?<addon>[^ \n]+)\s*(?<cached>⚡|❌)?

url
`)).MustParse()
