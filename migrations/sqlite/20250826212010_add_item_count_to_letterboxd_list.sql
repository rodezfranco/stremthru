-- +goose Up
-- +goose StatementBegin
ALTER TABLE `letterboxd_list` ADD COLUMN `item_count` int NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE `letterboxd_list` DROP COLUMN `item_count`;
-- +goose StatementEnd
