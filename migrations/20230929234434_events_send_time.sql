-- +goose Up
-- +goose StatementBegin
ALTER TABLE events
    ADD COLUMN send_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ADD COLUMN sended BOOLEAN NOT NULL DEFAULT FALSE;
UPDATE events
SET    send_time = start,
       sended = (notification -> 'sended')::BOOLEAN,
       notification  = to_jsonb(notification -> 'notification_params');
ALTER TABLE events
    RENAME COLUMN notification to notification_params;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE events set notification_params = json_build_object('sended', sended, 'notification_params', notification_params);
ALTER TABLE events
    DROP COLUMN send_time,
    DROP COLUMN sended;
ALTER TABLE events
    RENAME COLUMN notification_params to notification;
-- +goose StatementEnd
