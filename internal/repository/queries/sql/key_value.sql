-- name: GetValue :one
SELECT value
FROM key_value
WHERE key = @key;

-- name: SetValue :exec
INSERT INTO key_value
(key, value)
VALUES (@key, @value)
ON CONFLICT (key)
DO UPDATE
SET value = @value;

-- name: DeleteValue :exec
DELETE
FROM key_value
WHERE key = @key;
