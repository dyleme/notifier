-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS tasks;
ALTER TABLE basic_events RENAME TO basic_tasks;
ALTER TABLE periodic_events RENAME TO periodic_tasks;

ALTER TABLE notifications RENAME COLUMN event_id TO task_id;
ALTER TABLE notifications RENAME COLUMN event_type TO task_type;

ALTER TYPE event_type RENAME TO task_type;

ALTER TYPE task_type RENAME VALUE 'basic_event' TO 'basic_task';
ALTER TYPE task_type RENAME VALUE 'periodic_event' TO 'periodic_task';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TYPE task_type RENAME VALUE 'basic_task' TO 'basic_event';
ALTER TYPE task_type RENAME VALUE 'periodic_task' TO 'periodic_event';

ALTER TYPE task_type RENAME TO event_type;

ALTER TABLE notifications RENAME COLUMN task_id TO event_id;
ALTER TABLE notifications RENAME COLUMN task_type TO event_type;

ALTER TABLE basic_tasks RENAME TO basic_events;
ALTER TABLE periodic_tasks RENAME TO periodic_events;

CREATE TABLE tasks (
    id          SERIAL                                      PRIMARY KEY,
    created_at  TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW()   NOT NULL,
    message     CHARACTER VARYING(250)                      NOT NULL,
    user_id     INTEGER                                     NOT NULL,
    periodic    BOOLEAN DEFAULT FALSE                       NOT NULL,
    archived    BOOLEAN DEFAULT FALSE                       NOT NULL
);



-- +goose StatementEnd
