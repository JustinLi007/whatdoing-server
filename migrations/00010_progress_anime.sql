-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS progress_anime (
  id UUID PRIMARY KEY,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  episode INT NOT NULL DEFAULT 0,
  CONSTRAINT gt_zero CHECK ( episode >= 0 ),
  anime_id UUID NOT NULL,
  CONSTRAINT fk_anime_id FOREIGN KEY (anime_id) REFERENCES anime (id) ON DELETE CASCADE,
  user_library_id UUID NOT NULL,
  CONSTRAINT fk_user_library_id FOREIGN KEY (user_library_id) REFERENCES user_library (id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE progress_anime;
-- +goose StatementEnd
