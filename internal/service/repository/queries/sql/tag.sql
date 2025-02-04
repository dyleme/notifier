-- name: AddTag :one
INSERT INTO 
tags (name, user_id) 
VALUES (@name, @user_id)
RETURNING *;

-- name: GetTag :one
SELECT * FROM tags
WHERE id = @id;

-- name: ListTags :many
SELECT * FROM tags
WHERE user_id = @user_id
LIMIT @lim 
OFFSET @off;

-- name: DeleteTag :exec
DELETE FROM tags
WHERE id = @id;


-- name: AddTagsToSmth :copyfrom
INSERT INTO
smth2tags (smth_id, tag_id, user_id)
VALUES (@smth_id, @tag_id, @user_id);


-- name: ListTagsForSmth :many
SELECT t.* FROM smth2tags as s2t
JOIN tags as t 
ON s2t.tag_id = t.id
WHERE smth_id = @smth_id;

-- name: AmountOfExistedTags :one
SELECT COUNT(*)
FROM tags
WHERE tag_id = ANY(@tag_ids::int[])
  AND user_id = @user_id;

-- name: ListTagsForSmths :many
SELECT s2t.smth_id,sqlc.embed(t) FROM smth2tags as s2t
JOIN tags as t 
ON s2t.tag_id = t.id
WHERE smth_id = ANY(@smth_ids::int[]);

-- name: DeleteTagsFromSmth :exec
DELETE FROM smth2tags
WHERE smth_id = @smth_id 
AND tag_id = ANY(@tag_ids::int[]);

-- name: DeleteAllTagsForSmth :exec
DELETE FROM smth2tags
WHERE smth_id = @smth_id;

-- name: UpdateTag :exec
UPDATE tags
SET name = @name
WHERE id = @id;
