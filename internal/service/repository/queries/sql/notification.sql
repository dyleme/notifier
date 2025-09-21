-- name: ListNotSentNotifications :many
SELECT e.id as event_id, e.next_sending, t.text, u.tg_id, u.notification_retry_period_s
FROM events AS e
JOIN tasks AS t
    ON e.task_id = t.id
JOIN users AS u
    ON t.user_id = u.id 
WHERE e.done = 0
  AND e.next_sending <= @till
ORDER BY e.next_sending DESC;
