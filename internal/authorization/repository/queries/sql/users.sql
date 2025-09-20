-- name: AddUser :one
INSERT INTO users (
    tg_id
)
VALUES (
    @tg_id
) RETURNING *;

-- name: GetUser :one
SELECT * 
FROM users
WHERE id = @id;

-- name: FindUserByTgID :one
SELECT *
FROM users
WHERE tg_id = @tg_id;

-- name: UpdateUser :exec
UPDATE users
SET timezone_offset = @timezone_offset,
    timezone_dst = @timezone_dst
WHERE tg_id = @tg_id;
