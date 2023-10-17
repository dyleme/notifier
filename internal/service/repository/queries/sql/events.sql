-- name: AddEvent :one
INSERT INTO events (user_id,
                    text,
                    start,
                    description,
                    done,
                    notification_params,
                    send_time,
                    sended)
VALUES (@user_id,
        @text,
        @start,
        @description,
        @done,
        @notification_params,
        @send_time,
        @sended)
RETURNING *;

-- name: GetEvent :one
SELECT *
FROM events
WHERE id = @id
  AND user_id = @user_id;

-- name: ListEvents :many
SELECT *
FROM events
WHERE user_id = @user_id
ORDER BY id DESC
LIMIT @lim OFFSET @OFF;

-- name: CountListEvents :one
SELECT COUNT(*)
FROM events
WHERE user_id = @user_id;


-- name: GetEventsInPeriod :many
SELECT *
FROM events
WHERE user_id = @user_id
  AND start BETWEEN @from_time AND @to_time
ORDER BY id DESC
LIMIT @lim OFFSET @off;


-- name: CountGetEventsInPeriod :one
SELECT COUNT(*)
FROM events
WHERE user_id = @user_id
  AND start BETWEEN @from_time AND @to_time;

-- name: DeleteEvent :many
DELETE
FROM events
WHERE id = @id
AND user_id = @user_id
RETURNING *;

-- name: UpdateEvent :one
UPDATE events
SET start       = @start,
    text        = @text,
    description = @description,
    done        = @done,
    notification_params = @notification_params,
    sended = @sended,
    send_time = @send_time
WHERE id = @id
  AND user_id = @user_id
RETURNING *;

-- name: NearestEventTime :one
SELECT send_time as t
FROM events
WHERE done = FALSE
  AND sended = FALSE
ORDER BY start
LIMIT 1;

-- name: ListNearestEvents :many
SELECT *
FROM events
WHERE sended = FALSE
  AND send_time < @nearest_time
ORDER BY start;

-- name: MarkSendedNotificationEvent :exec
UPDATE events
SET sended = TRUE
WHERE id = @event_id;

-- name: UpdateNotificationParams :one
UPDATE events AS t
SET notification_params = notification_params ||  @params
WHERE id = @id
  AND user_id = @user_id
RETURNING notification_params;

-- name: DelayEvent :exec
UPDATE events AS t
SET send_time =  sqlc.arg(till)::TIMESTAMP,
    sended = FALSE
WHERE id = @id
  AND user_id = @user_id;
