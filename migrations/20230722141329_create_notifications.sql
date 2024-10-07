-- +goose Up
-- +goose StatementBegin
ALTER TABLE timetable_tasks
ADD COLUMN
notification jsonb
DEFAULT '{"sended":false}';
CREATE TABLE IF NOT EXISTS default_user_notification_params
(
    user_id integer PRIMARY KEY,
    created_at timestamp NOT NULL DEFAULT NOW(),
    params jsonb,
    CONSTRAINT fk_notifications_params_user_id FOREIGN KEY (
        user_id
    ) REFERENCES users
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS default_user_notification_params;
ALTER TABLE timetable_tasks
DROP COLUMN
notification;
-- +goose StatementEnd
