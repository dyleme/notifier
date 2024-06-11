-- name: AddBasicTask :one
INSERT INTO basic_tasks (
  user_id,
  text,
  start,
  description,
  notification_params
) VALUES (
  @user_id,
  @text,
  @start,
  @description,
  @notification_params
)
RETURNING *;

-- name: GetBasicTask :one
SELECT *
FROM basic_tasks
WHERE id = @id;

-- name: ListBasicTasks :many
SELECT *
FROM basic_tasks
WHERE user_id = @user_id
ORDER BY id DESC
LIMIT @lim OFFSET @OFF;

-- name: CountListBasicTasks :one
SELECT COUNT(*)
FROM basic_tasks
WHERE user_id = @user_id;


-- name: DeleteBasicTask :many
DELETE
FROM basic_tasks
WHERE id = @id
RETURNING *;

-- name: UpdateBasicTask :one
UPDATE basic_tasks
SET start       = @start,
    text        = @text,
    description = @description,
    notification_params = @notification_params
WHERE id = @id
  AND user_id = @user_id
RETURNING *;
