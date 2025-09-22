

-- +goose Up
-- +goose StatementBegin
CREATE VIEW events AS
SELECT 
    t.id as task_id,
    s.id as sending_id,
    s.done,
    s.original_sending,
    s.next_sending,
    t.text, 
    t.description, 
    u.tg_id, 
    u.id as user_id,
    u.notification_retry_period_s
FROM sendings AS s
JOIN tasks AS t
    ON s.task_id = t.id
JOIN users AS u
    ON t.user_id = u.id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW events;
-- +goose StatementEnd
