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
WHERE id = @id
  AND user_id = @user_id;

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


-- name: ListPeriodicEventsInPeriod :many
SELECT *
FROM periodic_events AS pe
         JOIN periodic_events_notifications AS pen
              ON pe.id = pen.periodic_event_id
WHERE user_id = @user_id
  AND pen.send_time BETWEEN @from_time AND @to_time
ORDER BY pen.send_time DESC
LIMIT @lim OFFSET @OFF;


-- name: CountListPeriodicEventsInPeriod :one
SELECT COUNT(*)
FROM periodic_events AS pe
         JOIN periodic_events_notifications AS pen
              ON pe.id = pen.periodic_event_id
WHERE user_id = @user_id
  AND pen.send_time BETWEEN @from_time AND @to_time;

-- name: DeletePeriodicEvent :many
DELETE
FROM periodic_events
WHERE id = @id
  AND user_id = @user_id
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

-- name: AddPeriodicEventNotification :one
INSERT INTO periodic_events_notifications
(periodic_event_id,
 send_time
)
VALUES (@periodic_event_id,
        @send_time
       )
RETURNING *;

-- name: UpdatePeriodicEventNotification :exec
UPDATE periodic_events_notifications
SET sended = @sended,
    send_time = @send_time
WHERE id = @id
  AND periodic_event_id = @periodic_event_id;

-- name: CurrentPeriodicEventNotification :one
SELECT *
FROM periodic_events_notifications
WHERE periodic_event_id = @periodic_event_id
  AND sended = FALSE
ORDER BY send_time DESC
LIMIT 1;

-- name: DelayPeriodicEventNotification :exec
UPDATE periodic_events_notifications
SET send_time = sqlc.arg(till)::TIMESTAMP,
    sended    = FALSE
WHERE id = @id;

-- name: NearestPeriodicEventTime :one
SELECT send_time AS t
FROM periodic_events_notifications
WHERE done = FALSE
  AND sended = FALSE
ORDER BY send_time
LIMIT 1;

-- name: ListNearestPeriodicEvents :many
SELECT *
FROM periodic_events AS pe
         JOIN periodic_events_notifications AS pen
              ON pe.id = pen.periodic_event_id
WHERE pen.done = FALSE
  AND pen.sended = FALSE
  AND pen.send_time < @nearest_time
ORDER BY pen.send_time;

-- name: DeletePeriodicEventNotification :many
DELETE FROM periodic_events_notifications
WHERE id = @id
  AND periodic_event_id = @periodic_event_id
RETURNING *;