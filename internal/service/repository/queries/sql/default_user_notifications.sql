-- name: SetDefaultUserNotificationParams :one
INSERT INTO default_user_notification_params (user_id,
                                              params
)
VALUES (@user_id,
        @params
       )
ON CONFLICT (user_id)
    DO UPDATE SET params          = @params
RETURNING *;

-- name: GetDefaultUserNotificationParams :one
SELECT *
FROM default_user_notification_params
WHERE user_id = @user_id;