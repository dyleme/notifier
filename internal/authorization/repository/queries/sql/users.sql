-- name: AddUser :one
INSERT INTO users (
    tg_id
)
VALUES (
    @tg_id
) RETURNING *;

-- name: FindUser :one
SELECT *
FROM users
WHERE tg_id = @tg_id;

-- -- name: AddBindingAttempt :exec
-- INSERT INTO binding_attempts (
--     tg_id,
--     code,
--     password_hash
-- ) VALUES (
--     @tg_id,
--     @code,
--     @password_hash
-- );

-- -- name: GetLatestBindingAttempt :one
-- SELECT *
-- FROM binding_attempts
-- WHERE tg_id = @tg_id
-- ORDER BY login_timestamp DESC
-- LIMIT 1;

-- -- name: UpdateBindingAttempt :exec
-- UPDATE binding_attempts
-- SET done = @done
-- WHERE id = @id;

-- -- name: GetLoginParameters :one
-- SELECT id,
--        password_hash
-- FROM users
-- WHERE tg_nickname = @tg_nickname;

-- name: UpdateUser :exec
UPDATE users
SET timezone_offset = @timezone_offset,
    timezone_dst = @timezone_dst
WHERE tg_id = @tg_id;
