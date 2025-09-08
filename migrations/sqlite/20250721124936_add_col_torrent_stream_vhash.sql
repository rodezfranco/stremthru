-- +goose Up
-- +goose StatementBegin
ALTER TABLE `torrent_stream` ADD COLUMN `vhash` varchar NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `torrent_stream` DROP COLUMN `vhash`;
-- +goose StatementEnd
