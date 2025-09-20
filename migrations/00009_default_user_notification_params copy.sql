
-- +goose Up
-- +goose StatementBegin
CREATE TABLE default_user_notification_params (
    user_id INTEGER PRIMARY KEY,
    params JSON NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE default_user_notification_params;
-- +goose StatementEnd
