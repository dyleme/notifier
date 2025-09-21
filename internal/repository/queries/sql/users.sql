-- name: GetUser :one
SELECT * FROM users
WHERE id = @id;

-- name: CreateUser :one
INSERT INTO users (
    tg_id,
    timezone_offset,
    timezone_dst,
    notification_retry_period_s
) VALUES (
    ?,?,?,?
) RETURNING *;

-- name: GetUserByTgID :one
SELECT * FROM users
WHERE tg_id = @tg_id;

-- name: UpdateUser :exec
UPDATE users
SET
    timezone_offset = @timezone_offset,
    timezone_dst = @timezone_dst,
    notification_retry_period_s = @notification_retry_period_s
WHERE id = @id;
