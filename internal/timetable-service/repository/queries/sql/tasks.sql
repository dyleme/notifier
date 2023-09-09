-- name: AddTask :one
INSERT INTO tasks (user_id,
                   message,
                   required_time)
VALUES (@user_id,
        @message,
        @required_time)
RETURNING *;

-- name: GetTask :one
SELECT *
FROM tasks
WHERE id = @id
  AND user_id = @user_id;

-- name: UpdateTask :exec
UPDATE tasks
SET required_time = @required_time,
    message       = @message,
    periodic      = @periodic,
    done          = @done,
    archived      = @archived
WHERE id = @id
  AND user_id = @user_id;

-- name: DeleteTask :execrows
DELETE
FROM tasks
WHERE id = @id
  AND user_id = @user_id;

-- name: ListUserTasks :many
SELECT *
FROM tasks
WHERE user_id = @user_id
  AND archived = FALSE;