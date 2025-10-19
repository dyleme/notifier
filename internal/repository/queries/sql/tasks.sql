-- name: AddTask :one
INSERT INTO tasks (
  text,
  description,
  user_id,
  type,
  start,
  event_creation_params
) VALUES (
  ?,?,?,?,time(?),?
) RETURNING *;

-- name: GetTask :one
SELECT *
FROM tasks
WHERE id      = ?
  AND user_id = ?;

-- name: ListTasks :many
SELECT *
FROM tasks
WHERE user_id = ?
  AND type = ?
ORDER BY id DESC
LIMIT ? OFFSET ?;

-- name: CountListSingleTasks :one
SELECT COUNT(*)
FROM tasks;

-- name: DeleteTask :many
DELETE
FROM tasks
WHERE id      = ?
  AND user_id = ?
RETURNING *;

-- name: UpdateTask :exec
UPDATE tasks
SET text        = ?,
    description = ?,
    event_creation_params = ?,
    start = time(?)
WHERE id = ?
  AND user_id = ?;