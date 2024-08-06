-- name: AddPeriodicTask :one
INSERT INTO periodic_tasks (user_id,
                             text,
                             start,
                             smallest_period,
                             biggest_period,
                             description,
                             notification_params
)
VALUES (@user_id,
        @text,
        @start,
        @smallest_period,
        @biggest_period,
        @description,
        @notification_params)
RETURNING *;

-- name: GetPeriodicTask :one
SELECT *
FROM periodic_tasks
WHERE id = @id;

-- name: ListPeriodicTasks :many
SELECT sqlc.embed(pt)
FROM periodic_tasks as pt
LEFT JOIN smth_to_tags as s2t
  ON pt.id = s2t.smth_id
LEFT JOIN tags as t
  ON s2t.tag_id = t.id
WHERE pt.user_id = @user_id
  AND (
    t.id = ANY (@tag_ids::int[]) 
    OR array_length(@tag_ids::int[], 1) is null
  )
ORDER BY pt.id DESC
LIMIT @lim OFFSET @OFF;

-- name: CountListPeriodicTasks :one
SELECT COUNT(*)
FROM periodic_tasks as pt
LEFT JOIN smth_to_tags as s2t
  ON pt.id = s2t.smth_id
LEFT JOIN tags as t
  ON s2t.tag_id = t.id
WHERE pt.user_id = @user_id
  AND (
    t.id = ANY (@tag_ids::int[]) 
    OR array_length(@tag_ids::int[], 1) is null
  );


-- name: DeletePeriodicTask :many
DELETE
FROM periodic_tasks
WHERE id = @id
RETURNING *;

-- name: UpdatePeriodicTask :one
UPDATE periodic_tasks
SET start               = @start,
    text                = @text,
    description         = @description,
    notification_params = @notification_params,
    smallest_period     = @smallest_period,
    biggest_period      = @biggest_period
WHERE id = @id
  AND user_id = @user_id
RETURNING *;