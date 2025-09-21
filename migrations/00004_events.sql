
-- +goose Up
-- +goose StatementBegin
CREATE TABLE events
(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    task_id INTEGER NOT NULL,
    next_sending DATETIME NOT NULL,
    original_sending DATETIME NOT NULL,
    done SMALLINT NOT NULL DEFAULT 0
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE events;
-- +goose StatementEnd
