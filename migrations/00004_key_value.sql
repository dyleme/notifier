
-- +goose Up
-- +goose StatementBegin
CREATE TABLE key_value (
    key VARCHAR(100) PRIMARY KEY,
    value JSON NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE key_value;
-- +goose StatementEnd
