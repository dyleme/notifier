-- name: AddUser :one
INSERT INTO users (
                   email,
                   password_hash,
                   tg_id
                   )
VALUES (
        @email,
        @password_hash,
        @tg_id
       )
RETURNING *;

-- name: FindUser :one
SELECT *
FROM users
WHERE email = @email
   OR tg_id = @tg_id;

-- name: GetLoginParameters :one
SELECT id,
       password_hash
FROM users
WHERE email = @email;