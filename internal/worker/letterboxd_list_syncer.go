package worker

import (
	"net/http"
	"time"

	"github.com/rodezfranco/stremthru/internal/config"
	"github.com/rodezfranco/stremthru/internal/db"
	"github.com/rodezfranco/stremthru/internal/letterboxd"
	"github.com/rodezfranco/stremthru/internal/peer"
	"github.com/rodezfranco/stremthru/internal/util"
	"github.com/rodezfranco/stremthru/internal/worker/worker_queue"
)

func InitSyncLetterboxdList(conf *WorkerConfig) *Worker {
	conf.Executor = func(w *Worker) error {
		log := w.Log

		worker_queue.LetterboxdListSyncerQueue.Process(func(item worker_queue.LetterboxdListSyncerQueueItem) error {
			l, err := letterboxd.GetListById(item.ListId)
			if err != nil {
				return err
			}

			if l == nil {
				log.Warn("list not found in database", "id", item.ListId)
				return nil
			}

			if !l.IsStale() && !l.HasUnfetchedItems() {
				log.Debug("list already synced", "id", item.ListId, "name", l.Name)
				return nil
			}

			if !config.Integration.Letterboxd.IsEnabled() {
				if !config.HasPeer {
					return nil
				}

				if !util.HasDurationPassedSince(l.UpdatedAt.Time, 15*time.Minute) {
					return worker_queue.ErrWorkerQueueItemDelayed
				}

				log.Debug("fetching list by id from upstream", "id", l.Id)
				res, err := Peer.FetchLetterboxdList(&peer.FetchLetterboxdListParams{
					ListId: l.Id,
				})
				if err != nil {
					return err
				}

				list := &res.Data

				l.UserId = list.UserId
				l.UserName = list.UserSlug
				l.Name = list.Title
				l.Slug = list.Slug
				l.Description = list.Description
				l.Private = list.IsPrivate
				l.ItemCount = list.ItemCount
				l.Items = nil
				for i := range list.Items {
					item := &list.Items[i]
					l.Items = append(l.Items, letterboxd.LetterboxdItem{
						Id:          item.Id,
						Name:        item.Title,
						ReleaseYear: item.Year,
						Runtime:     item.Runtime,
						Rating:      item.Rating,
						Adult:       item.IsAdult,
						Poster:      item.Poster,
						UpdatedAt:   db.Timestamp{Time: item.UpdatedAt},

						GenreIds: item.GenreIds,
						IdMap:    &item.IdMap,
						Rank:     item.Index,
					})
				}

				if err := letterboxd.UpsertList(l); err != nil {
					return err
				}

				letterboxd.InvalidateListCache(l)

				return nil
			}

			items := []letterboxd.LetterboxdItem{}

			hasMore := true
			perPage := 100
			page := 0
			cursor := ""
			for hasMore {
				page++
				log.Debug("fetching list items", "id", l.Id, "page", page)
				res, err := letterboxd.Client.FetchListEntries(&letterboxd.FetchListEntriesParams{
					Id:      l.Id,
					Cursor:  cursor,
					PerPage: perPage,
				})
				if err != nil {
					if res.StatusCode == http.StatusTooManyRequests {
						duration := letterboxd.Client.GetRetryAfter()
						log.Warn("rate limited, cooling down", "duration", duration, "id", l.Id, "page", page)
						time.Sleep(duration)
						page--
						continue
					}
					log.Error("failed to fetch list items", "error", err, "id", l.Id, "page", page)
					return err
				}

				now := time.Now()
				for i := range res.Data.Items {
					item := &res.Data.Items[i]
					rank := item.Rank
					if rank == 0 {
						rank = i
					}
					items = append(items, letterboxd.LetterboxdItem{
						Id:          item.Film.Id,
						Name:        item.Film.Name,
						ReleaseYear: item.Film.ReleaseYear,
						Runtime:     item.Film.RunTime,
						Rating:      int(item.Film.Rating * 2 * 10),
						Adult:       item.Film.Adult,
						Poster:      item.Film.GetPoster(),
						UpdatedAt:   db.Timestamp{Time: now},

						GenreIds: item.Film.GenreIds(),
						IdMap:    item.Film.GetIdMap(),
						Rank:     rank,
					})
				}

				cursor = res.Data.Next
				hasMore = cursor != "" && len(res.Data.Items) == perPage
				time.Sleep(5 * time.Second)
			}

			l.Items = items

			if err := letterboxd.UpsertList(l); err != nil {
				return err
			}

			letterboxd.InvalidateListCache(l)

			return nil
		})

		return nil

	}

	worker := NewWorker(conf)

	return worker
}
