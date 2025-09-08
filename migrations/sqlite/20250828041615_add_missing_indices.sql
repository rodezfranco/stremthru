-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS `anidb_torrent_idx_hash` ON `anidb_torrent` (`hash`);

CREATE INDEX IF NOT EXISTS `imdb_title_map_idx_lboxd` ON `imdb_title_map` (`lboxd`);
CREATE INDEX IF NOT EXISTS `imdb_title_map_idx_tmdb` ON `imdb_title_map` (`tmdb`);
CREATE INDEX IF NOT EXISTS `imdb_title_map_idx_trakt` ON `imdb_title_map` (`trakt`);
CREATE INDEX IF NOT EXISTS `imdb_title_map_idx_tvdb` ON `imdb_title_map` (`tvdb`);

CREATE INDEX IF NOT EXISTS `torrent_info_idx_parser_version` ON `torrent_info` (`parser_version`);

CREATE INDEX IF NOT EXISTS `torrent_stream_idx_asid` ON `torrent_stream` (`asid`);
CREATE INDEX IF NOT EXISTS `torrent_stream_idx_asid_nocase` ON `torrent_stream` (`asid` COLLATE nocase);
CREATE INDEX IF NOT EXISTS `torrent_stream_idx_sid` ON `torrent_stream` (`sid`);
CREATE INDEX IF NOT EXISTS `torrent_stream_idx_sid_nocase` ON `torrent_stream` (`sid` COLLATE nocase);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS `torrent_stream_idx_sid_nocase`;
DROP INDEX IF EXISTS `torrent_stream_idx_sid`;
DROP INDEX IF EXISTS `torrent_stream_idx_asid_nocase`;
DROP INDEX IF EXISTS `torrent_stream_idx_asid`;

DROP INDEX IF EXISTS `torrent_info_idx_parser_version`;

DROP INDEX IF EXISTS `imdb_title_map_idx_tvdb`;
DROP INDEX IF EXISTS `imdb_title_map_idx_trakt`;
DROP INDEX IF EXISTS `imdb_title_map_idx_tmdb`;
DROP INDEX IF EXISTS `imdb_title_map_idx_lboxd`;

DROP INDEX IF EXISTS `anidb_torrent_idx_hash`;
-- +goose StatementEnd
