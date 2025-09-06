-- +goose Up
-- +goose StatementBegin
-- +goose StatementEnd
CREATE TABLE smth2tags (
  smth_id INTEGER NOT NULL,
  tag_id INTEGER NOT NULL,
  user_id INTEGER NOT NULL,
  PRIMARY KEY (smth_id),
  UNIQUE (smth_id, tag_id),
  FOREIGN KEY (tag_id, user_id) REFERENCES tags(id, user_id)
)
-- +goose Down
-- +goose StatementBegin
DROP TABLE smth2tags;
-- +goose StatementEnd
