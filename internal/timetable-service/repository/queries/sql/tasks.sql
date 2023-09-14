-- name: AddTask :one
INSERT INTO tasks (user_id,
                   message,
                   periodic )
VALUES (@user_id,
        @message,
        @periodic)
RETURNING *;

-- name: GetTask :one
SELECT *
FROM tasks
WHERE id = @id
  AND user_id = @user_id;

-- name: UpdateTask :exec
UPDATE tasks
SET message       = @message,
    periodic      = @periodic,
    archived      = @archived
WHERE id = @id
  AND user_id = @user_id;

-- name: DeleteTask :execrows
DELETE
FROM tasks
WHERE id = @id
  AND user_id = @user_id;

-- name: ListTasks :many
SELECT *
FROM tasks
WHERE user_id = @user_id
  AND archived = FALSE
ORDER BY id DESC
LIMIT @lim
OFFSET @off;


-- name: CountListTasks :one
SELECT count(*)
FROM tasks
WHERE user_id = @user_id
  AND archived = FALSE;
