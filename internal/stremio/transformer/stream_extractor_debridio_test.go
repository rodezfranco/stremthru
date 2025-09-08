package stremio_transformer

import (
	"testing"

	"github.com/MunifTanjim/go-ptt"
	"github.com/MunifTanjim/stremthru/stremio"
	"github.com/stretchr/testify/assert"
)

func TestStreamExtractorDebridioDebrid(t *testing.T) {
	for _, tc := range []struct {
		name   string
		sType  string
		stream stremio.Stream
		result StreamExtractorResult
	}{
		{
			"single",
			"movie", stremio.Stream{
				Name:  "[ED ⚡] \nDebridio 1080p",
				Title: "Deadpool 2016 1080p BluRay x264 DTS-JYK\n⚡ 📺 1080p 💾 2.72 GB  ⚙️ ThePirateBay",
				URL:   "https://addon.debridio.com/play/movie/easydebrid/xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx/xxxxxxxxxxxxxxxx/0f61c0478326c8e2f8a397f59d7917a0dc558718/Deadpool%202016%201080p%20BluRay%20x264%20DTS-JYK",
			}, StreamExtractorResult{
				Hash:   "0f61c0478326c8e2f8a397f59d7917a0dc558718",
				TTitle: "Deadpool 2016 1080p BluRay x264 DTS-JYK",
				Result: &ptt.Result{
					Resolution: "1080p",
					Site:       "ThePirateBay",
					Size:       "2.72 GB",
				},
				Addon: StreamExtractorResultAddon{
					Name: "Debridio",
				},
				File: StreamExtractorResultFile{
					Idx: -1,
				},
				Store: StreamExtractorResultStore{
					Name:     "easydebrid",
					Code:     "ED",
					IsCached: true,
				},
				Episode: -1,
				Season:  -1,
			},
		},
		{
			"multi",
			"series", stremio.Stream{
				Name:  "[ED ⚡] \nDebridio 4k",
				Title: "Reacher.S03E07.Si.va.a.Los.Angeles.ITA.ENG.2160p.AMZN.WEB-DL.DDP5.1.DV.HDR.H.265-MeM.GP.mkv\n⚡ 📺 4k 💾 5.55 GB  ⚙️ RARBG\n🌐 🇬🇧|🇮🇹",
				URL:   "https://addon.debridio.com/play/serie/easydebrid/04745965c0043769872b248542624fdf/7tsnerwp6jdpuixz/76a1fa8d28d4c04b201803ab7262d037d295134f/tt9288030/3/7/Reacher.S03E07.Si.va.a.Los.Angeles.ITA.ENG.2160p.AMZN.WEB-DL.DDP5.1.DV.HDR.H.265-MeM.GP.mkv",
				BehaviorHints: &stremio.StreamBehaviorHints{
					BingeGroup: "debridio-4k",
					Filename:   "Reacher.S03E07.Si.va.a.Los.Angeles.ITA.ENG.2160p.AMZN.WEB-DL.DDP5.1.DV.HDR.H.265-MeM.GP.mkv",
				},
			}, StreamExtractorResult{
				Hash:   "76a1fa8d28d4c04b201803ab7262d037d295134f",
				TTitle: "Reacher.S03E07.Si.va.a.Los.Angeles.ITA.ENG.2160p.AMZN.WEB-DL.DDP5.1.DV.HDR.H.265-MeM.GP.mkv",
				Result: &ptt.Result{
					Episodes:   []int{7},
					Languages:  []string{"en", "it"},
					Resolution: "4k",
					Seasons:    []int{3},
					Site:       "RARBG",
					Size:       "5.55 GB",
				},
				Addon: StreamExtractorResultAddon{
					Name: "Debridio",
				},
				File: StreamExtractorResultFile{
					Idx:  -1,
					Name: "Reacher.S03E07.Si.va.a.Los.Angeles.ITA.ENG.2160p.AMZN.WEB-DL.DDP5.1.DV.HDR.H.265-MeM.GP.mkv",
				},
				Store: StreamExtractorResultStore{
					Name:     "easydebrid",
					Code:     "ED",
					IsCached: true,
				},
				Episode: 7,
				Season:  3,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := StreamExtractorDebridio.Parse(&tc.stream, tc.sType)
			tc.result.Category = tc.sType
			tc.result.Result.Normalize()
			tc.result.Raw.Name = tc.stream.Name
			tc.result.Raw.Description = tc.stream.Title
			assert.Equal(t, &tc.result, data)
		})
	}
}
