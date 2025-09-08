package worker

import (
	"slices"
	"sync"
	"time"

	"github.com/rodezfranco/stremthru/internal/db"
	"github.com/rodezfranco/stremthru/internal/imdb_title"
	"github.com/rodezfranco/stremthru/internal/imdb_torrent"
	"github.com/rodezfranco/stremthru/internal/torrent_info"
)

func InitMapIMDBTorrentWorker(conf *WorkerConfig) *Worker {
	conf.Executor = func(w *Worker) error {
		if !isIMDBSyncedToday() {
			w.Log.Info("IMDB not synced yet today, skipping")
			return nil
		}

		batch_size := 10000
		chunk_size := 1000
		if db.Dialect == db.DBDialectPostgres {
			batch_size = 20000
			chunk_size = 2000
		}

		totalCount := 0
		for {
			hashes, err := torrent_info.GetIMDBUnmappedHashes(batch_size)
			if err != nil {
				return err
			}

			var wg sync.WaitGroup
			for cHashes := range slices.Chunk(hashes, chunk_size) {
				wg.Add(1)
				go func() {
					defer wg.Done()

					items := []imdb_torrent.IMDBTorrent{}
					tInfoByHash, err := torrent_info.GetByHashes(cHashes)
					if err != nil {
						w.Log.Error("failed to get torrent info", "error", err)
						return
					}
					hashesByCategory := map[torrent_info.TorrentInfoCategory][]string{
						torrent_info.TorrentInfoCategoryMovie:  {},
						torrent_info.TorrentInfoCategorySeries: {},
					}
					for hash, tInfo := range tInfoByHash {
						if !tInfo.IsParsed() {
							continue
						}

						ito := imdb_torrent.IMDBTorrent{
							Hash: hash,
						}

						if tInfo.Title == "" {
							items = append(items, ito)
							continue
						}

						titleType := imdb_title.SearchTitleTypeUnknown
						if tInfo.Category == torrent_info.TorrentInfoCategoryMovie {
							titleType = imdb_title.SearchTitleTypeMovie
							hashesByCategory[torrent_info.TorrentInfoCategoryMovie] = append(hashesByCategory[torrent_info.TorrentInfoCategoryMovie], hash)
						} else if tInfo.Category == torrent_info.TorrentInfoCategorySeries || len(tInfo.Seasons) > 0 || len(tInfo.Episodes) > 0 {
							titleType = imdb_title.SearchTitleTypeShow
							hashesByCategory[torrent_info.TorrentInfoCategorySeries] = append(hashesByCategory[torrent_info.TorrentInfoCategorySeries], hash)
						} else if tInfo.Category == torrent_info.TorrentInfoCategoryXXX {
							// ¯\_(ツ)_/¯
						} else {
							titleType = imdb_title.SearchTitleTypeMovie
							hashesByCategory[torrent_info.TorrentInfoCategoryMovie] = append(hashesByCategory[torrent_info.TorrentInfoCategoryMovie], hash)
						}

						imdbTitle, err := imdb_title.SearchOne(tInfo.Title, titleType, tInfo.Year, false)
						if err != nil {
							w.Log.Error("failed to search imdb title", "error", err, "title", tInfo.Title, "year", tInfo.Year)
							continue
						}
						if imdbTitle != nil {
							ito.TId = imdbTitle.TId
						}
						items = append(items, ito)
					}

					if err := imdb_torrent.Insert(items); err != nil {
						w.Log.Error("failed to map imdb torrent", "error", err)
						return
					}
					torrent_info.SetMissingCategory(hashesByCategory)

					w.Log.Info("mapped imdb torrent", "count", len(items))
				}()
			}
			wg.Wait()

			count := len(hashes)
			totalCount += count
			w.Log.Info("processed torrents", "totalCount", totalCount)

			if count < batch_size {
				break
			}

			time.Sleep(200 * time.Millisecond)
		}

		return nil
	}

	return NewWorker(conf)
}
