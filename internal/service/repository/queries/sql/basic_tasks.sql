-- name: AddBasicTask :one
INSERT INTO single_tasks (
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
FROM single_tasks
WHERE id = @id;

-- name: ListBasicTasks :many
SELECT sqlc.embed(bt)
FROM single_tasks as bt
LEFT JOIN smth2tags as s2t
  ON bt.id = s2t.smth_id
LEFT JOIN tags as t
  ON s2t.tag_id = t.id
WHERE bt.user_id = @user_id
  AND (
    @tag_ids IS NULL 
    OR t.id IN (SELECT value FROM json_each(@tag_ids))
  )
ORDER BY bt.id DESC
LIMIT @lim OFFSET @off;

-- name: CountListBasicTasks :one
SELECT COUNT(*)
FROM single_tasks as bt
LEFT JOIN smth2tags as s2t
  ON bt.id = s2t.smth_id
LEFT JOIN tags as t
  ON s2t.tag_id = t.id
WHERE bt.user_id = @user_id
  AND (
    @tag_ids IS NULL 
    OR t.id IN (SELECT value FROM json_each(@tag_ids))
  );

-- name: DeleteBasicTask :many
DELETE
FROM single_tasks
WHERE id = @id
RETURNING *;

-- name: UpdateBasicTask :exec
UPDATE single_tasks
SET start       = @start,
    text        = @text,
    description = @description,
    notification_params = @notification_params
WHERE id = @id
  AND user_id = @user_id;
