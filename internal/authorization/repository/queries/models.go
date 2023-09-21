// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.19.1

package queries

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type DefaultUserNotificationParam struct {
	UserID    int32
	CreatedAt pgtype.Timestamp
	Params    []byte
}

type Event struct {
	ID           int32
	CreatedAt    pgtype.Timestamp
	Text         string
	Description  pgtype.Text
	UserID       int32
	Start        pgtype.Timestamptz
	Done         bool
	Notification []byte
}

type Task struct {
	ID        int32
	CreatedAt pgtype.Timestamp
	Message   string
	UserID    int32
	Periodic  bool
	Archived  bool
}

type User struct {
	ID             int32
	Email          pgtype.Text
	PasswordHash   pgtype.Text
	TgID           pgtype.Int4
	TimezoneOffset int32
	TimezoneDst    bool
}
