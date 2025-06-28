-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS rel_anime_anime_names (
  id UUID PRIMARY KEY,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  anime_id UUID NOT NULL,
  CONSTRAINT fk_anime_id
  FOREIGN KEY (anime_id)
  REFERENCES anime (id)
  ON DELETE CASCADE,
  anime_names_id UUID NOT NULL,
  CONSTRAINT fk_anime_names_id
  FOREIGN KEY (anime_names_id)
  REFERENCES anime_names (id)
  ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE rel_anime_anime_names;
-- +goose StatementEnd
