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
SELECT *
FROM periodic_tasks
WHERE user_id = @user_id
ORDER BY id DESC
LIMIT @lim OFFSET @OFF;

-- name: CountListPeriodicTasks :one
SELECT COUNT(*)
FROM periodic_tasks
WHERE user_id = @user_id;


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