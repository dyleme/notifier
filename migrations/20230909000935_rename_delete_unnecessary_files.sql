-- +goose Up
-- +goose StatementBegin
ALTER TABLE timetable_tasks RENAME TO events;
ALTER TABLE events DROP CONSTRAINT fk_events_user;
ALTER TABLE events DROP COLUMN task_id;
ALTER TABLE events DROP COLUMN finish;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events RENAME TO timetable_tasks;
ALTER TABLE timetable_tasks ADD COLUMN task_id INTEGER;
ALTER TABLE timetable_tasks ADD CONSTRAINT fk_events_tasks FOREIGN KEY (task_id) REFERENCES tasks;
-- +goose StatementEnd
