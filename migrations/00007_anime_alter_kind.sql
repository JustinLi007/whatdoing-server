-- +goose Up
-- +goose StatementBegin
ALTER TABLE anime
ADD COLUMN kind TEXT NOT NULL DEFAULT 'anime';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE anime DROP COLUMN kind;
-- +goose StatementEnd
