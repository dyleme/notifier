-- name: AddBasicEvent :one
INSERT INTO basic_events (
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

-- name: GetBasicEvent :one
SELECT *
FROM basic_events
WHERE id = @id;

-- name: ListBasicEvents :many
SELECT *
FROM basic_events
WHERE user_id = @user_id
ORDER BY id DESC
LIMIT @lim OFFSET @OFF;

-- name: CountListBasicEvents :one
SELECT COUNT(*)
FROM basic_events
WHERE user_id = @user_id;


-- name: DeleteBasicEvent :many
DELETE
FROM basic_events
WHERE id = @id
RETURNING *;

-- name: UpdateBasicEvent :one
UPDATE basic_events
SET start       = @start,
    text        = @text,
    description = @description,
    notification_params = @notification_params
WHERE id = @id
  AND user_id = @user_id
RETURNING *;
