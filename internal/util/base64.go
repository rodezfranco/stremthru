package util

import "github.com/rodezfranco/stremthru/core"

func MustDecodeBase64(value string) string {
	blob, err := core.Base64Decode(value)
	if err != nil {
		panic(err)
	}
	return string(blob)
}
