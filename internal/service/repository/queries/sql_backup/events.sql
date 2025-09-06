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
) ;

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
    @tag_ids IS NULL 
    OR t.id IN (SELECT value FROM json_each(@tag_ids))
  )
ORDER BY next_send DESC
LIMIT @lim OFFSET @off;

-- name: DeleteEvent :many
DELETE FROM events
WHERE id = @id
;

-- name: UpdateEvent :one
UPDATE events
SET text = @text,
    next_send = @next_send,
    first_send = @first_send,
    done = @done
WHERE id = @id
;

-- name: ListNotSendedEvents :many
SELECT * FROM events
WHERE next_send <= @till
  AND done = 0
  AND notify = 1;

-- name: GetNearestEventTime :one
SELECT next_send FROM events
WHERE done = 0
  AND notify = 1 
ORDER BY next_send ASC
LIMIT 1;

-- name: ListUserDailyEvents :many
SELECT * FROM events
WHERE user_id = @user_id
  AND datetime(next_send, @time_offset) BETWEEN date('now') AND date('now', '+1 day')
ORDER BY next_send ASC;

-- name: ListNotDoneEvents :many
SELECT * FROM events
WHERE user_id = @user_id
  AND done = 0
  AND next_send < datetime('now')
ORDER BY next_send ASC;
