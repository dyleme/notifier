-- name: AddEvent :one
INSERT INTO events (user_id,
                             text,
                             start,
                             description,
                             done,
                             notification)
VALUES (@user_id,
        @text,
        @start,
        @description,
        @done,
        @notification)
RETURNING *;

-- name: GetEvent :one
SELECT *
FROM events
WHERE id = @id
  AND user_id = @user_id;

-- name: ListEvents :many
SELECT *
FROM events
WHERE user_id = @user_id;

-- name: GetEventsInPeriod :many
SELECT *
FROM events
WHERE user_id = @user_id
  AND start BETWEEN @from_time AND @to_time;

-- name: DeleteEvent :one
DELETE
FROM events
WHERE id = @id
  AND user_id = @user_id
RETURNING COUNT(*) AS deleted_amount;

-- name: UpdateEvent :one
UPDATE events
SET start       = @start,
    text        = @text,
    description = @description,
    done        = @done
WHERE id = @id
  AND user_id = @user_id
RETURNING *;

-- name: GetEventReadyTasks :many
SELECT *
FROM events AS t
WHERE t.start <= NOW()
  AND t.done = FALSE
  AND t.notification ->> 'sended' = 'false'
  AND (
        t.notification ->> 'delayed_till' IS NULL
        OR CAST(t.notification ->> 'delayed_till' AS TIMESTAMP) <= NOW()
    );

-- name: MarkNotificationSended :exec
UPDATE events AS t
SET notification = notification || '{"sended":true}'
WHERE id = ANY (sqlc.arg(ids)::INTEGER[]);

-- name: UpdateNotificationParams :one
UPDATE events AS t
SET notification = JSONB_SET(notification, '{notification_params}', @params)
WHERE id = @id
  AND user_id = @user_id
RETURNING notification;

-- name: DelayEvent :exec
UPDATE events AS t
SET notification = JSONB_SET(notificaiton, '{"delayed_till"}',sqlc.arg(till)::TIMESTAMP)
WHERE id = @id
  AND user_id = @user_id;
