
-- +goose Up
-- +goose StatementBegin
CREATE TABLE single_tasks
(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    text VARCHAR(250) NOT NULL,
    description VARCHAR(250),
    user_id INTEGER NOT NULL,
    start DATETIME NOT NULL,
    notification_params JSON,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE single_tasks;
-- +goose StatementEnd
