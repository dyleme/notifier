-- +goose Up
-- +goose StatementBegin
CREATE TABLE key_value (
  key varchar(100) NOT NULL UNIQUE,
  value JSONB NOT NULL
);
CREATE INDEX key_value_key ON key_value USING HASH(key);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX key_value_key;
DROP TABLE key_value;
-- +goose StatementEnd
