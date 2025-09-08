package stremio_shared

import (
	"strings"

	"github.com/MunifTanjim/stremthru/internal/util"
	"github.com/MunifTanjim/stremthru/stremio"
)

var claimedManifestIdPrefixForStremioAddonsDotNet = util.MustDecodeBase64("eHl6LjEzMzc3MDAxLnN0cmVtdGhydQ==")

func ClaimAddonOnStremioAddonsDotNet(manifest *stremio.Manifest, signature string) {
	if !strings.HasPrefix(manifest.ID, claimedManifestIdPrefixForStremioAddonsDotNet) {
		return
	}

	if manifest.StremioAddonsConfig == nil {
		manifest.StremioAddonsConfig = &stremio.StremioAddonsConfig{}
	}

	manifest.StremioAddonsConfig.Issuer = "https://stremio-addons.net"
	manifest.StremioAddonsConfig.Signature = signature
}
