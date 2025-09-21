-- name: AddEvent :one
INSERT INTO events (
  task_id,
  done,
  original_sending,
  next_sending
) VALUES (
  ?,?,?,?
) RETURNING *;

-- name: GetEvent :one
SELECT * FROM events
WHERE id = @id;

-- name: GetLatestEvent :one
SELECT * FROM events
WHERE task_id = @task_id
ORDER BY next_send DESC
LIMIT 1;
  
-- name: ListUserEvents :many
SELECT DISTINCT e.*
FROM events as e
JOIN tasks as t
ON e.task_id = t.id
WHERE t.user_id = ?
  AND next_sending <= @to_time
  AND next_sending >= @from_time
ORDER BY next_sending DESC
LIMIT ? OFFSET ?;

-- name: DeleteEvent :many
DELETE FROM events
WHERE id = @id
RETURNING *;

-- name: UpdateEvent :one
UPDATE events
SET
  next_sending     = ?,
  original_sending = ?,
  done             = ?
WHERE 
  id = ?
RETURNING *;

-- name: ListNotSendedEvents :many
SELECT * FROM events
WHERE next_sending <= @till
  AND done = 0
  AND notify = 1;

-- name: GetNearestEventTime :one
SELECT next_sending FROM events
WHERE done = 0
  AND notify = 1 
ORDER BY next_sending ASC
LIMIT 1;

-- name: RescheduleEvent :exec
UPDATE events
SET next_sending = ?
WHERE id = ?;