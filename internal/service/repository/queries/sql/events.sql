-- name: AddEvent :one
INSERT INTO events (
    user_id,
    text,
    task_id,
    task_type,
    next_send, 
    notification_params,
    first_send
) VALUES (
    @user_id,
    @text,
    @task_id,
    @task_type,
    @next_send,
    @notification_params,
    @next_send
) RETURNING *;

-- name: GetEvent :one
SELECT * FROM events
WHERE id = @id;

-- name: GetLatestEvent :one
SELECT * FROM events
WHERE task_id = @task_id
  AND task_type = @task_type
ORDER BY next_send DESC
LIMIT 1;
  
-- name: ListUserEvents :many
SELECT DISTINCT sqlc.embed(e) 
FROM events as e
LEFT JOIN smth2tags as s2t
  ON e.id = s2t.smth_id
LEFT JOIN tags as t
  ON s2t.tag_id = t.id
WHERE e.user_id = @user_id
  AND next_send BETWEEN @from_time AND @to_time
  AND (
    t.id = ANY (@tag_ids::int[]) 
    OR array_length(@tag_ids::int[], 1) is null
  )
ORDER BY next_send DESC
LIMIT @lim OFFSET @off;

-- name: DeleteEvent :many
DELETE FROM events
WHERE id = @id
RETURNING *;

-- name: UpdateEvent :one
UPDATE events
SET text = @text,
    next_send = @next_send,
    first_send = @first_send,
    done = @done
WHERE id = @id
RETURNING *;

-- name: ListNotSendedEvents :many
SELECT * FROM events
WHERE next_send <= @till
  AND done = false
  AND notify = true;

-- name: GetNearestEventTime :one
SELECT next_send FROM events
WHERE done = false
  AND notify = true 
ORDER BY next_send ASC
LIMIT 1;

-- name: ListUserDailyEvents :many
SELECT * FROM events
WHERE user_id = @user_id
  AND next_send + @time_offset  BETWEEN CURRENT_DATE AND CURRENT_DATE + 1
ORDER BY next_send ASC;

-- name: ListNotDoneEvents :many
SELECT * FROM events
WHERE user_id = @user_id
  AND done = false
  AND next_send < NOW()
ORDER BY next_send ASC;
