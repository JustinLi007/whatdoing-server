-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS rel_anime_user_library (
  id UUID PRIMARY KEY,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  status TEXT NOT NULL DEFAULT 'not-started',
  CONSTRAINT valid_status CHECK (status = 'started' OR status = 'not-started' OR status = 'completed'),
  episode INT NOT NULL DEFAULT 1,
  CONSTRAINT gt_zero CHECK ( episode > 0 ),
  anime_id UUID NOT NULL,
  CONSTRAINT fk_anime_id FOREIGN KEY (anime_id) REFERENCES anime (id) ON DELETE CASCADE,
  user_library_id UUID NOT NULL,
  CONSTRAINT fk_user_library_id FOREIGN KEY (user_library_id) REFERENCES user_library (id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE rel_anime_user_library;
-- +goose StatementEnd
