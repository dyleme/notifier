-- +goose Up
-- +goose StatementBegin
ALTER TABLE events
    ALTER COLUMN start TYPE TIMESTAMPTZ;
ALTER TABLE users
    ADD COLUMN timezone_offset INTEGER NOT NULL DEFAULT +3,
    ADD COLUMN timezone_dst BOOLEAN NOT NULL DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events
    ALTER COLUMN start TYPE TIMESTAMP;
ALTER TABLE users
    DROP COLUMN timezone_offset,
    DROP COLUMN timezone_dst;
-- +goose StatementEnd
