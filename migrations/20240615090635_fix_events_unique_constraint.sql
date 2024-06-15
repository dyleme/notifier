-- +goose Up
-- +goose StatementBegin
ALTER TABLE events DROP CONSTRAINT notifications_event_id_event_type_key;
CREATE INDEX events_task_id_task_type_idx ON events (task_id, task_type);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events ADD CONSTRAINT notifications_event_id_event_type_key UNIQUE(task_id, task_type);
DROP INDEX events_task_id_task_type_idx;
-- +goose StatementEnd
