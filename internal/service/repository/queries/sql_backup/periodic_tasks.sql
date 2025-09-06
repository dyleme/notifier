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
;

-- name: GetPeriodicTask :one
SELECT *
FROM periodic_tasks
WHERE id = @id;

-- name: ListPeriodicTasks :many
SELECT sqlc.embed(pt)
FROM periodic_tasks as pt
LEFT JOIN smth2tags as s2t
  ON pt.id = s2t.smth_id
LEFT JOIN tags as t
  ON s2t.tag_id = t.id
WHERE pt.user_id = @user_id
  AND (
    @tag_ids IS NULL 
    OR t.id IN (SELECT value FROM json_each(@tag_ids))
  )
ORDER BY pt.id DESC
LIMIT @lim OFFSET @OFF;

-- name: CountListPeriodicTasks :one
SELECT COUNT(*)
FROM periodic_tasks as pt
LEFT JOIN smth2tags as s2t
  ON pt.id = s2t.smth_id
LEFT JOIN tags as t
  ON s2t.tag_id = t.id
WHERE pt.user_id = @user_id
  AND (
    @tag_ids IS NULL 
    OR t.id IN (SELECT value FROM json_each(@tag_ids))
  );

-- name: DeletePeriodicTask :many
DELETE
FROM periodic_tasks
WHERE id = @id
;

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
;
