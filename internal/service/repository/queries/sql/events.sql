-- name: AddEvent :one
INSERT INTO events (
    user_id,
    text,
    task_id,
    task_type,
    time, 
    notification_params,
    first_time,
    last_sended_time
) VALUES (
    @user_id,
    @text,
    @task_id,
    @task_type,
    @time,
    @notification_params,
    @time,
    TIMESTAMP '1970-01-01 00:00:00'
) RETURNING *;

-- name: GetEvent :one
SELECT * FROM events
WHERE id = @id;

-- name: GetLatestEvent :one
SELECT * FROM events
WHERE task_id = @task_id
  AND task_type = @task_type
ORDER BY time DESC
LIMIT 1;
  
-- name: ListUserEvents :many
SELECT DISTINCT sqlc.embed(e) 
FROM events as e
LEFT JOIN smth_to_tags as s2t
  ON e.id = s2t.smth_id
LEFT JOIN tags as t
  ON s2t.tag_id = t.id
WHERE e.user_id = @user_id
  AND time BETWEEN @from_time AND @to_time
  AND (
    t.id = ANY (@tag_ids::int[]) 
    OR array_length(@tag_ids::int[], 1) is null
  )
ORDER BY time DESC
LIMIT @lim OFFSET @off;

-- name: DeleteEvent :many
DELETE FROM events
WHERE id = @id
RETURNING *;

-- name: UpdateEvent :one
UPDATE events
SET text = @text,
    time = @time,
    first_time = @first_time,
    done = @done
WHERE id = @id
RETURNING *;

-- name: ListNotSendedEvents :many
SELECT * FROM events
WHERE time <= @till
  AND done = false
  AND notify = true;

-- name: GetNearestEventTime :one
SELECT time FROM events
WHERE done = false
  AND notify = true 
ORDER BY time ASC
LIMIT 1;

-- name: ListUserDailyEvents :many
SELECT * FROM events
WHERE user_id = @user_id
  AND time + @time_offset  BETWEEN CURRENT_DATE AND CURRENT_DATE + 1
ORDER BY time ASC;

-- name: ListNotDoneEvents :many
SELECT * FROM events
WHERE user_id = @user_id
  AND done = false
  AND time < NOW()
ORDER BY time ASC;