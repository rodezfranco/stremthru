package stremio_wrap

import (
	"errors"
	"strings"

	"github.com/rodezfranco/stremthru/internal/config"
	"github.com/rodezfranco/stremthru/internal/kv"
	stremio_transformer "github.com/rodezfranco/stremthru/internal/stremio/transformer"
	"github.com/rodezfranco/stremthru/stremio"
)

type StreamTransformer struct {
	Extractor stremio_transformer.StreamExtractor
	Template  *stremio_transformer.StreamTemplate
}

type WrappedStream struct {
	*stremio.Stream
	r              *stremio_transformer.StreamExtractorResult
	noContentProxy bool
}

func (ws WrappedStream) IsSortable() bool {
	return ws.r != nil
}

func (ws WrappedStream) GetQuality() string {
	return ws.r.Quality
}

func (ws WrappedStream) GetResolution() string {
	return ws.r.Resolution
}

func (ws WrappedStream) GetSize() string {
	return ws.r.Size
}

func (ws WrappedStream) GetHDR() string {
	return strings.Join(ws.r.HDR, "|")
}

func (st StreamTransformer) Do(stream *stremio.Stream, sType string, tryReconfigure bool) (*WrappedStream, error) {
	s := &WrappedStream{Stream: stream}

	if st.Template == nil || st.Template.IsEmpty() {
		return s, nil
	}

	data := st.Extractor.Parse(stream, sType)
	if data == nil {
		return s, nil
	}

	if tryReconfigure {
		if s.URL != "" && data.Hash != "" {
			s.InfoHash = data.Hash
			s.FileIndex = data.File.Idx
			s.URL = ""
			data.Store.Code = ""
			data.Store.Name = ""
			data.Store.IsCached = false
			if data.File.Name != "" {
				if s.BehaviorHints == nil {
					s.BehaviorHints = &stremio.StreamBehaviorHints{}
				}
				if s.BehaviorHints.Filename == "" {
					s.BehaviorHints.Filename = data.File.Name
				}
			}
		}
	}

	s.r = data

	var err error
	s.Stream, err = st.Template.Execute(s.Stream, data)
	if err != nil {
		return s, err
	}

	return s, nil
}

const BUILTIN_TRANSFORMER_ENTITY_ID_EMOJI = "✨"
const BUILTIN_TRANSFORMER_ENTITY_ID_PREFIX = BUILTIN_TRANSFORMER_ENTITY_ID_EMOJI + " "

var newTransformerExtractorIdMap = map[string]string{
	"Debridio":    BUILTIN_TRANSFORMER_ENTITY_ID_PREFIX + "Debridio",
	"Mediafusion": BUILTIN_TRANSFORMER_ENTITY_ID_PREFIX + "MediaFusion",
	"Torrentio":   BUILTIN_TRANSFORMER_ENTITY_ID_PREFIX + "Torrentio",
}

func getNewTransformerExtractorId(oldId string) string {
	if newId, ok := newTransformerExtractorIdMap[oldId]; ok {
		return newId
	}
	return oldId
}

var builtInExtractors = func() map[string]stremio_transformer.StreamExtractorBlob {
	extractors := map[string]stremio_transformer.StreamExtractorBlob{}

	extractors[BUILTIN_TRANSFORMER_ENTITY_ID_PREFIX+"Comet"] = stremio_transformer.StreamExtractorComet.Blob
	extractors[BUILTIN_TRANSFORMER_ENTITY_ID_PREFIX+"Debridio"] = stremio_transformer.StreamExtractorDebridio.Blob
	extractors[BUILTIN_TRANSFORMER_ENTITY_ID_PREFIX+"MediaFusion"] = stremio_transformer.StreamExtractorMediaFusion.Blob
	extractors[BUILTIN_TRANSFORMER_ENTITY_ID_PREFIX+"Peerflix"] = stremio_transformer.StreamExtractorPeerflix.Blob
	extractors[BUILTIN_TRANSFORMER_ENTITY_ID_PREFIX+"Torrentio"] = stremio_transformer.StreamExtractorTorrentio.Blob
	extractors[BUILTIN_TRANSFORMER_ENTITY_ID_PREFIX+"Orion"] = stremio_transformer.StreamExtractorOrion.Blob
	extractors[BUILTIN_TRANSFORMER_ENTITY_ID_PREFIX+"Cinema"] = stremio_transformer.StreamExtractorCinema.Blob

	return extractors
}()

var extractorStore = kv.NewKVStore[stremio_transformer.StreamExtractorBlob](&kv.KVStoreConfig{
	Type: "st:wrap:transformer:extractor",
	GetKey: func(key string) string {
		return key
	},
})

func getExtractor(extractorId string) (stremio_transformer.StreamExtractorBlob, error) {
	if strings.HasPrefix(extractorId, BUILTIN_TRANSFORMER_ENTITY_ID_EMOJI) {
		if extractor, ok := builtInExtractors[extractorId]; ok {
			return extractor, nil
		}
		return "", errors.New("built-in extractor not found")
	}

	var extractor stremio_transformer.StreamExtractorBlob
	if err := extractorStore.GetValue(extractorId, &extractor); err != nil {
		return "", err
	}
	return extractor, nil
}

func getExtractorIds() ([]string, error) {
	extractors, err := extractorStore.List()
	if err != nil {
		return nil, err
	}
	builtInExtractorsCount := len(builtInExtractors)
	extractorIds := make([]string, builtInExtractorsCount+len(extractors))
	idx := 0
	for id := range builtInExtractors {
		extractorIds[idx] = id
		idx++
	}
	for _, extractor := range extractors {
		extractorIds[idx] = extractor.Key
		idx++
	}
	return extractorIds, nil

}

var builtInTemplates = func() map[string]stremio_transformer.StreamTemplateBlob {
	templates := map[string]stremio_transformer.StreamTemplateBlob{}

	templates[BUILTIN_TRANSFORMER_ENTITY_ID_PREFIX+"Default"] = stremio_transformer.StreamTemplateDefault.Blob
	templates[BUILTIN_TRANSFORMER_ENTITY_ID_PREFIX+"Raw"] = stremio_transformer.StreamTemplateRaw.Blob
	templates[BUILTIN_TRANSFORMER_ENTITY_ID_PREFIX+"Cinema"] = stremio_transformer.StreamTemplateCinema.Blob
	return templates
}()

var templateStore = kv.NewKVStore[stremio_transformer.StreamTemplateBlob](&kv.KVStoreConfig{
	Type: "st:wrap:transformer:template",
	GetKey: func(key string) string {
		return key
	},
})

func getTemplate(templateId string) (stremio_transformer.StreamTemplateBlob, error) {
	if strings.HasPrefix(templateId, BUILTIN_TRANSFORMER_ENTITY_ID_EMOJI) {
		if template, ok := builtInTemplates[templateId]; ok {
			return template, nil
		}
		return stremio_transformer.StreamTemplateBlob{}, errors.New("built-in template not found")
	}

	var template stremio_transformer.StreamTemplateBlob
	if err := templateStore.GetValue(templateId, &template); err != nil {
		return stremio_transformer.StreamTemplateBlob{}, err
	}
	return template, nil
}

func getTemplateIds() ([]string, error) {
	templates, err := templateStore.List()
	if err != nil {
		return nil, err
	}
	builtInTemplatesCount := len(builtInTemplates)
	templateIds := make([]string, builtInTemplatesCount+len(templates))
	idx := 0
	for id := range builtInTemplates {
		templateIds[idx] = id
		idx++
	}
	for _, template := range templates {
		templateIds[idx] = template.Key
		idx++
	}
	return templateIds, nil
}

func seedDefaultTransformerEntities() {
	if config.IsPublicInstance {
		for oldId := range newTransformerExtractorIdMap {
			if err := extractorStore.Del(oldId); err != nil {
				log.Warn("Failed to cleanup seed extractor: " + oldId)
			}
		}
	}
	for id := range builtInExtractors {
		if err := extractorStore.Del(id); err != nil {
			log.Warn("Failed to cleanup seed extractor: " + id)
		}
	}

	for key := range builtInTemplates {
		if err := templateStore.Del(key); err != nil {
			log.Warn("Failed to cleanup seed template: " + key)
		}
		if config.IsPublicInstance {
			key = strings.TrimPrefix(key, BUILTIN_TRANSFORMER_ENTITY_ID_PREFIX)
			if err := templateStore.Del(key); err != nil {
				log.Warn("Failed to cleanup seed template: " + key)
			}
		}
	}
}
