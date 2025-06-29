-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS rel_users_anime (
  id UUID PRIMARY KEY,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  user_id UUID NOT NULL,
  CONSTRAINT fk_user_id
  FOREIGN KEY (user_id)
  REFERENCES users (id)
  ON DELETE CASCADE,
  anime_id UUID NOT NULL,
  CONSTRAINT fk_anime_id
  FOREIGN KEY (anime_id)
  REFERENCES anime (id)
  ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE rel_users_anime;
-- +goose StatementEnd
