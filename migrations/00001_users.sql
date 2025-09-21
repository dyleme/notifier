-- +goose Up 
-- +goose StatementBegin 
CREATE TABLE users ( 
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tg_id BIGINT NOT NULL,
    timezone_offset INT NOT NULL DEFAULT 0,
    timezone_dst SMALLINT NOT NULL DEFAULT 0,
    notification_retry_period_s INT NOT NULL
);

CREATE INDEX users_tg_id_idx ON users(tg_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd
