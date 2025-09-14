
-- +goose Up
-- +goose StatementBegin
CREATE TABLE events
(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    user_id INTEGER NOT NULL,
    text TEXT NOT NULL,
    description TEXT,
    task_id INTEGER NOT NULL,
    task_type TEXT NOT NULL,
    next_send DATETIME NOT NULL,
    first_send DATETIME NOT NULL,
    done SMALLINT NOT NULL DEFAULT 0,
    notify SMALLINT NOT NULL DEFAULT 1,
    notification_params JSON NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE events;
-- +goose StatementEnd
