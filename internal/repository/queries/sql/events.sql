-- name: ListNotSentEvents :many
SELECT * 
FROM events
WHERE done = 0
  AND next_sending <= @till
ORDER BY next_sending DESC;

-- name: GetEvent :one
SELECT *
FROM events
WHERE sending_id = ?
  AND user_id = ?;

-- name: ListEvents :many
SELECT *
FROM events
WHERE user_id=?
  AND next_sending >= @from_time
  AND next_sending <= @to_time
  AND done = false
ORDER BY next_sending DESC
LIMIT ? OFFSET ?;


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


-- name: UpdateSending :one
UPDATE sendings
SET
  next_sending     = ?,
  original_sending = ?,
  done             = ?
WHERE 
  id = ?
RETURNING *;

-- name: GetLatestSending :one
SELECT * FROM sendings
WHERE task_id = @task_id
  AND done = 0
ORDER BY next_sending DESC
LIMIT 1;

-- name: DeleteSending :many
DELETE FROM sendings
WHERE id = @id
RETURNING *;

-- name: GetNearestSendingTime :one
SELECT next_sending FROM sendings
WHERE done = 0
ORDER BY next_sending ASC
LIMIT 1;