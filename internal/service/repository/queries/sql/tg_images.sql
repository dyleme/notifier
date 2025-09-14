-- name: AddTgImage :one
INSERT INTO tg_images (
                    filename,
                    tg_file_id
                    )
VALUES (@filename,
        @tg_file_id)
RETURNING *;

-- name: GetTgImage :one
SELECT *
FROM tg_images
WHERE filename = @filename;
