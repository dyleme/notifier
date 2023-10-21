// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.1
// source: tg_images.sql

package queries

import (
	"context"
)

const addTgImage = `-- name: AddTgImage :one
INSERT INTO tg_images (
                    filename,
                    tg_file_id
                    )
VALUES ($1,
        $2)
RETURNING id, filename, tg_file_id
`

type AddTgImageParams struct {
	Filename string `db:"filename"`
	TgFileID string `db:"tg_file_id"`
}

func (q *Queries) AddTgImage(ctx context.Context, db DBTX, arg AddTgImageParams) (TgImage, error) {
	row := db.QueryRow(ctx, addTgImage, arg.Filename, arg.TgFileID)
	var i TgImage
	err := row.Scan(&i.ID, &i.Filename, &i.TgFileID)
	return i, err
}

const getTgImage = `-- name: GetTgImage :one
SELECT id, filename, tg_file_id
FROM tg_images
WHERE filename = $1
`

func (q *Queries) GetTgImage(ctx context.Context, db DBTX, filename string) (TgImage, error) {
	row := db.QueryRow(ctx, getTgImage, filename)
	var i TgImage
	err := row.Scan(&i.ID, &i.Filename, &i.TgFileID)
	return i, err
}
