-- +goose Up
-- +goose StatementBegin
CREATE TABLE users
(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tg_id BIGINT NOT NULL,
    timezone_offset INT NOT NULL DEFAULT 0,
    timezone_dst SMALLINT NOT NULL DEFAULT 0
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd
