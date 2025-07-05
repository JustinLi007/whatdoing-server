-- +goose Up
-- +goose StatementBegin
ALTER TABLE anime
ADD COLUMN image_url TEXT DEFAULT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE anime DROP COLUMN image_url;
-- +goose StatementEnd
