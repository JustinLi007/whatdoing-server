-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS jwt (
  id UUID PRIMARY KEY,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  token BYTEA NOT NULL UNIQUE,
  refresh_token BYTEA NOT NULL UNIQUE,
  refresh_token_expiration TIMESTAMP NOT NULL,
  scope TEXT NOT NULL,
  user_id UUID NOT NULL,
  CONSTRAINT fk_user_id
  FOREIGN KEY (user_id)
  REFERENCES users (id)
  ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE jwt;
-- +goose StatementEnd
