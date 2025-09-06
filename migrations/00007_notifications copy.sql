
-- +goose Up
-- +goose StatementBegin
CREATE TABLE notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    user_id INTEGER NOT NULL,
    text TEXT NOT NULL,
    description TEXT,
    event_id INTEGER NOT NULL,
    event_type TEXT NOT NULL,
    send_time DATETIME NOT NULL,
    sended BOOLEAN NOT NULL DEFAULT 0,
    done BOOLEAN NOT NULL DEFAULT 0,
    notification_params JSON,
    inline_keyboard TEXT,
    resend_count INTEGER DEFAULT 0,
    max_resend_count INTEGER DEFAULT 0,
    resend_interval TEXT,
    last_resend_time DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id),
    UNIQUE (event_id, event_type)
)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE notifications;
-- +goose StatementEnd
