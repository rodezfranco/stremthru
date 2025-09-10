package stremio_transformer

import "strings"

var StreamExtractorCinema = StreamExtractorBlob(strings.TrimSpace(`
name
(?i)^(?:Other|(?<resolution>\d[^kp]*[kp]))

description
^(?:🎬 (?<t_title>.+?))?(?: \((?<year>\d{4})\))?(?: 📗(?<season>\d+))?(?: 📘(?<episode>\d+))?(?: 🎥 (?<quality>.+?))?(?: 🎞 (?<encode>.+?))?(?: 🎧 (?<audioTags>[^🔊\n]+))?(?: 🔊 (?<audioChannels>.+?))?(?: 🌐 (?<languages>.+?))?(?: 📦 (?<size>.+?))?(?: ⏱ (?<duration>.+?))?(?: 🌱 (?<seeders>\d+))?(?: ℹ (?<message>.+?))?(?<addon>[^ \n]+)(?<cached>⚡|❌)?
`)).MustParse()
