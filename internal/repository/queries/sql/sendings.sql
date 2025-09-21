-- name: AddSending :one
INSERT INTO sendings (
  task_id,
  done,
  original_sending,
  next_sending
) VALUES (
  ?,?,?,?
) RETURNING *;

-- name: GetSendning :one
SELECT * FROM sendings
WHERE id = @id;

-- name: GetLatestSending :one
SELECT * FROM sendings
WHERE task_id = @task_id
ORDER BY next_send DESC
LIMIT 1;
  
-- name: ListUserSending :many
SELECT DISTINCT e.*
FROM sendings as e
JOIN tasks as t
ON e.task_id = t.id
WHERE t.user_id = ?
  AND next_sending <= @to_time
  AND next_sending >= @from_time
ORDER BY next_sending DESC
LIMIT ? OFFSET ?;

-- name: DeleteSending :many
DELETE FROM sendings
WHERE id = @id
RETURNING *;

-- name: UpdateSending :one
UPDATE sendings
SET
  next_sending     = ?,
  original_sending = ?,
  done             = ?
WHERE 
  id = ?
RETURNING *;

-- name: ListNotSendedSending :many
SELECT * FROM sendings
WHERE next_sending <= @till
  AND done = 0
  AND notify = 1;

-- name: GetNearestSendingTime :one
SELECT next_sending FROM sendings
WHERE done = 0
  AND notify = 1 
ORDER BY next_sending ASC
LIMIT 1;

-- name: RescheduleSending :exec
UPDATE sendings
SET next_sending = ?
WHERE id = ?;