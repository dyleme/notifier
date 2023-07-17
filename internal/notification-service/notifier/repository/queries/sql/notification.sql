-- name: AddNotification :exec
INSERT INTO notification (
            user_id,
            message,
            notification_time,
            destination,
            task_id
) VALUES (
    @user_id,
    @message,
    @notification_time,
    @destination,
    @task_id
);

-- name: GetNotification :one
SELECT *
  FROM notification
 WHERE id = @id;

-- name: FetchNewNotifications :many
SELECT *
  FROM notification
 WHERE notification_time < @till::timestamp
   AND sended = FALSE;

-- name: MarkSendedNotifications :exec
UPDATE notification 
   SET sended = TRUE
 WHERE notification_time < @till::timestamp;

-- name: GetFutureUserNotifications :many
SELECT *
  FROM notification
 WHERE notification_time > @from_time::timestamp
   AND user_id = @user_id::integer;

-- name: DeleteNotification :exec
DELETE
  FROM notification
 WHERE id = @id;
