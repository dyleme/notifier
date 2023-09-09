-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS timetable_tasks
(
    id          SERIAL PRIMARY KEY,
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW(),
    text        VARCHAR(250) NOT NULL,
    description VARCHAR(250),
    user_id     INTEGER      NOT NULL,
    start       TIMESTAMP    NOT NULL,
    finish      TIMESTAMP    NOT NULL,
    done        BOOLEAN      NOT NULL DEFAULT FALSE,
    task_id     INTEGER      NOT NULL,
    CONSTRAINT fk_events_user FOREIGN KEY (user_id) REFERENCES users,
    CONSTRAINT fk_events_tasks FOREIGN KEY (task_id) REFERENCES tasks
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS timetable_tasks;
-- +goose StatementEnd
