-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
    DROP COLUMN nickname,
    ADD COLUMN tg_id INTEGER UNIQUE,
    ALTER COLUMN email DROP NOT NULL,
    ALTER COLUMN password_hash DROP NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
    ADD COLUMN nickname VARCHAR(250),
    DROP COLUMN tg_id;
-- +goose StatementEnd
