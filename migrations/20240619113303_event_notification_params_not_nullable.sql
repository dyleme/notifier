-- +goose Up
-- +goose StatementBegin
UPDATE events AS ev
SET notification_params = dp.params
FROM default_user_notification_params AS dp
WHERE ev.user_id = dp.user_id;

ALTER TABLE events
ADD CONSTRAINT event_notification_params_not_nullable
CHECK (notification_params IS NOT NULL);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events
DROP CONSTRAINT event_notification_params_not_nullable;
-- +goose StatementEnd
