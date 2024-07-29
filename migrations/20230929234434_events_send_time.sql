-- +goose Up
-- +goose StatementBegin
ALTER TABLE events
ADD COLUMN send_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN sended BOOLEAN NOT NULL DEFAULT FALSE;
UPDATE events
SET
    send_time = start,
    sended = (notification -> 'sended')::BOOLEAN,
    notification = TO_JSONB(notification -> 'notification_params');
ALTER TABLE events
RENAME COLUMN notification TO notification_params;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE events SET
    notification_params = JSON_BUILD_OBJECT(
        'sended', sended,
        'notification_params', notification_params
    );
ALTER TABLE events
DROP COLUMN send_time,
DROP COLUMN sended;
ALTER TABLE events
RENAME COLUMN notification_params TO notification;
-- +goose StatementEnd
