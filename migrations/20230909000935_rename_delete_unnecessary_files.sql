-- +goose Up
-- +goose StatementBegin
ALTER TABLE timetable_tasks RENAME TO events;

ALTER TABLE events
DROP CONSTRAINT fk_events_user,
DROP COLUMN task_id,
DROP COLUMN finish;

ALTER TABLE tasks
DROP COLUMN done,
DROP COLUMN required_time;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events RENAME TO timetable_tasks;
ALTER TABLE timetable_tasks
ADD COLUMN task_id INTEGER,
ADD CONSTRAINT fk_events_tasks FOREIGN KEY (task_id) REFERENCES tasks;
-- +goose StatementEnd
