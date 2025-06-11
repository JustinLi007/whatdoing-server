-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS anime (
  id UUID PRIMARY KEY,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  name TEXT UNIQUE NOT NULL,
  episode INT NOT NULL DEFAULT 1,
  description TEXT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE anime;
-- +goose StatementEnd
