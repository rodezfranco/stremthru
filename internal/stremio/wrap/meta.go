package stremio_wrap

import (
	"net/http"

	"github.com/rodezfranco/stremthru/internal/context"
	"github.com/rodezfranco/stremthru/internal/shared"
	stremio_addon "github.com/rodezfranco/stremthru/internal/stremio/addon"
	"github.com/rodezfranco/stremthru/stremio"
)

func (ud UserData) fetchMeta(ctx *context.StoreContext, w http.ResponseWriter, r *http.Request, rType, id, extra string) error {
	upstreams, err := ud.getUpstreams(ctx, stremio.ResourceNameMeta, rType, id)
	if err != nil {
		return err
	}

	if len(upstreams) == 0 {
		shared.ErrorNotFound(r).Send(w, r)
		return nil
	}

	upstream := upstreams[0]

	addon.ProxyResource(w, r, &stremio_addon.ProxyResourceParams{
		BaseURL:  upstream.baseUrl,
		Resource: string(stremio.ResourceNameMeta),
		Type:     rType,
		Id:       id,
		Extra:    extra,
		ClientIP: ctx.ClientIP,
	})
	return nil
}
