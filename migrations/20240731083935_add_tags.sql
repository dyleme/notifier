-- +goose Up
-- +goose StatementBegin
CREATE TABLE tags (
  id SERIAL PRIMARY KEY,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  name VARCHAR(255) NOT NULL,
  user_id INT NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users(id),
  CONSTRAINT tass_unique_id_user_id UNIQUE (id, user_id),
  CONSTRAINT tags_unique_name_user_id UNIQUE (name, user_id)
);

CREATE TABLE smth2tags (
  smth_id INT NOT NULL,
  tag_id INT NOT NULL,
  user_id INT NOT NULL,
  PRIMARY KEY (smth_id),
  UNIQUE (smth_id, tag_id),
  FOREIGN KEY (tag_id, user_id) REFERENCES tags(id, user_id)
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE smth2tags;
DROP TABLE tags;
-- +goose StatementEnd
