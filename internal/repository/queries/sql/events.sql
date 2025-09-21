-- name: ListNotSentNotifications :many
SELECT s.id as event_id, s.next_sending, t.text, u.tg_id, u.notification_retry_period_s
FROM sendings AS s
JOIN tasks AS t
    ON s.task_id = t.id
JOIN users AS u
    ON t.user_id = u.id 
WHERE s.done = 0
  AND s.next_sending <= @till
ORDER BY s.next_sending DESC;
