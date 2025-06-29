-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS anime (
  id UUID PRIMARY KEY,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  episodes INT DEFAULT NULL,
  description TEXT DEFAULT NULL,
  anime_names_id UUID NOT NULL,
  CONSTRAINT fk_anime_names_id
  FOREIGN KEY (anime_names_id)
  REFERENCES anime_names (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE anime;
-- +goose StatementEnd
