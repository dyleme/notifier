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
SELECT sqlc.embed(bt)
FROM basic_tasks as bt
LEFT JOIN smth_to_tags as s2t
  ON bt.id = s2t.smth_id
LEFT JOIN tags as t
  ON s2t.tag_id = t.id
WHERE bt.user_id = @user_id
  AND (
    t.id = ANY (@tag_ids::int[]) 
    OR array_length(@tag_ids::int[], 1) is null
  )
ORDER BY bt.id DESC
LIMIT @lim OFFSET @OFF;

-- name: CountListBasicTasks :one
SELECT COUNT(*)
FROM basic_tasks as bt
LEFT JOIN smth_to_tags as s2t
  ON bt.id = s2t.smth_id
LEFT JOIN tags as t
  ON s2t.tag_id = t.id
WHERE bt.user_id = @user_id
  AND (
    t.id = ANY (@tag_ids::int[]) 
    OR array_length(@tag_ids::int[], 1) is null
  );


-- name: DeleteBasicTask :many
DELETE
FROM basic_tasks
WHERE id = @id
RETURNING *;

-- name: UpdateBasicTask :exec
UPDATE basic_tasks
SET start       = @start,
    text        = @text,
    description = @description,
    notification_params = @notification_params
WHERE id = @id
  AND user_id = @user_id;
