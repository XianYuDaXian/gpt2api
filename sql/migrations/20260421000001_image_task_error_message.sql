-- +goose Up
-- +goose StatementBegin

ALTER TABLE `image_tasks`
    ADD COLUMN `error_message` VARCHAR(1000) NOT NULL DEFAULT '' AFTER `error`;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE `image_tasks`
    DROP COLUMN `error_message`;

-- +goose StatementEnd
