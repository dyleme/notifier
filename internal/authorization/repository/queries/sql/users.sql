-- name: AddUser :one
INSERT INTO users (email,
                   nickname,
                   password_hash
                   )
VALUES (
        @email,
        @nickname,
        @password_hash
       )
RETURNING id;

-- name: GetLoginParameters :one
SELECT id,
       password_hash
FROM users
WHERE email = @auth_name
   OR nickname = @auth_name;