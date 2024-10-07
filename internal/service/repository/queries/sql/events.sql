-- name: AddEvent :one
INSERT INTO events (
    user_id,
    text,
    task_id,
    task_type,
    next_send_time, 
    notification_params,
    first_send_time,
    last_sended_time
) VALUES (
    @user_id,
    @text,
    @task_id,
    @task_type,
    @next_send_time,
    @notification_params,
    @next_send_time,
    TIMESTAMP '1970-01-01 00:00:00'
) RETURNING *;

-- name: GetEvent :one
SELECT * FROM events
WHERE id = @id;

-- name: GetLatestEvent :one
SELECT * FROM events
WHERE task_id = @task_id
  AND task_type = @task_type
ORDER BY next_send_time DESC
LIMIT 1;
  
-- name: ListUserEvents :many
SELECT * FROM events
WHERE user_id = @user_id
  AND next_send_time BETWEEN @from_time AND @to_time
ORDER BY next_send_time DESC
LIMIT @lim OFFSET @off;

-- name: DeleteEvent :many
DELETE FROM events
WHERE id = @id
RETURNING *;

-- name: UpdateEvent :one
UPDATE events
SET text = @text,
    next_send_time = @next_send_time,
    first_send_time = @first_send_time,
    done = @done
WHERE id = @id
RETURNING *;

-- name: ListNotSendedEvents :many
SELECT * FROM events
WHERE next_send_time <= @till
  AND done = false;

-- name: GetNearestEvent :one
SELECT * FROM events
WHERE done = false
ORDER BY next_send_time ASC
LIMIT 1;
