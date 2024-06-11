-- name: AddNotification :one
INSERT INTO notifications (
    user_id,
    text,
    task_id,
    task_type,
    send_time
) VALUES (
    @user_id,
    @text,
    @task_id,
    @task_type,
    @send_time
) RETURNING *;

-- name: GetNotification :one
SELECT * FROM notifications
WHERE id = @id;

-- name: GetLatestNotification :one
SELECT * FROM notifications
WHERE task_id = @task_id
ORDER BY send_time DESC
LIMIT 1;
  
-- name: ListUserNotifications :many
SELECT * FROM notifications
WHERE user_id = @user_id
  AND send_time BETWEEN @from_time AND @to_time
ORDER BY send_time DESC
LIMIT @lim OFFSET @off;

-- name: DeleteNotification :many
DELETE FROM notifications
WHERE id = @id
RETURNING *;

-- name: UpdateNotification :one
UPDATE notifications
SET text = @text,
    send_time = @send_time,
    sended = @sended,
    done = @done
WHERE id = @id
RETURNING *;

-- name: ListNotSendedNotifications :many
SELECT * FROM notifications
WHERE sended = FALSE
  AND send_time <= @till;

-- name: GetNearestNotification :one
SELECT * FROM notifications
WHERE sended = FALSE
  AND send_time <= @till
ORDER BY send_time ASC
LIMIT 1;

-- name: MarkSendedNotifiatoins :exec
UPDATE notifications
SET sended = TRUE
WHERE id = ANY(sqlc.slice(ids));
