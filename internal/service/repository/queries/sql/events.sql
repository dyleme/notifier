-- name: AddEvent :one
INSERT INTO events (
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

-- name: GetEvent :one
SELECT * FROM events
WHERE id = @id;

-- name: GetLatestEvent :one
SELECT * FROM events
WHERE task_id = @task_id
  AND task_type = @task_type
ORDER BY send_time DESC
LIMIT 1;
  
-- name: ListUserEvents :many
SELECT * FROM events
WHERE user_id = @user_id
  AND send_time BETWEEN @from_time AND @to_time
ORDER BY send_time DESC
LIMIT @lim OFFSET @off;

-- name: DeleteEvent :many
DELETE FROM events
WHERE id = @id
RETURNING *;

-- name: UpdateEvent :one
UPDATE events
SET text = @text,
    send_time = @send_time,
    sended = @sended,
    done = @done
WHERE id = @id
RETURNING *;

-- name: ListNotSendedEvents :many
SELECT * FROM events
WHERE sended = FALSE
  AND send_time <= @till;

-- name: GetNearestEvent :one
SELECT * FROM events
WHERE sended = FALSE
  AND send_time <= @till
ORDER BY send_time ASC
LIMIT 1;

-- name: MarkSendedNotifiatoins :exec
UPDATE events
SET sended = TRUE
WHERE id = ANY(sqlc.slice(ids));
