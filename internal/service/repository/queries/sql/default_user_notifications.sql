-- name: SetDefaultUserNotificationParams :one
INSERT INTO default_user_notification_params (user_id,
                                              params
)
VALUES (?1,
        ?2
       )
ON CONFLICT (user_id)
    DO UPDATE SET params          = ?2
RETURNING *;

-- name: GetDefaultUserNotificationParams :one
SELECT *
FROM default_user_notification_params
WHERE user_id = @user_id;
