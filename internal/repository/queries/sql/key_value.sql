-- name: GetValue :one
SELECT value
FROM key_value
WHERE key = @key;

-- name: SetValue :exec
INSERT INTO key_value
(key, value)
VALUES (?1, ?2)
ON CONFLICT (key)
DO UPDATE
SET value = ?2;

-- name: DeleteValue :exec
DELETE
FROM key_value
WHERE key = ?;
