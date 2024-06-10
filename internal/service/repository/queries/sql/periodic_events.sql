-- name: AddPeriodicEvent :one
INSERT INTO periodic_events (user_id,
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

-- name: GetPeriodicEvent :one
SELECT *
FROM periodic_events
WHERE id = @id;

-- name: ListPeriodicEvents :many
SELECT *
FROM periodic_events
WHERE user_id = @user_id
ORDER BY id DESC
LIMIT @lim OFFSET @OFF;

-- name: CountListPeriodicEvents :one
SELECT COUNT(*)
FROM periodic_events
WHERE user_id = @user_id;


-- name: DeletePeriodicEvent :many
DELETE
FROM periodic_events
WHERE id = @id
RETURNING *;

-- name: UpdatePeriodicEvent :one
UPDATE periodic_events
SET start               = @start,
    text                = @text,
    description         = @description,
    notification_params = @notification_params,
    smallest_period     = @smallest_period,
    biggest_period      = @biggest_period
WHERE id = @id
  AND user_id = @user_id
RETURNING *;