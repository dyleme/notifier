
-- +goose Up
-- +goose StatementBegin
CREATE TABLE tasks
(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    text TEXT NOT NULL,
    description TEXT NOT NULL,
    user_id INTEGER NOT NULL,
    type TEXT NOT NULL,
    start TEXT NOT NULL,
    event_creation_params JSONB NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE tasks;
-- +goose StatementEnd
