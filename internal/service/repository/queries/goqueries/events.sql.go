// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: events.sql

package goqueries

import (
	"context"

	domain "github.com/Dyleme/Notifier/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
)

const addEvent = `-- name: AddEvent :one
INSERT INTO events (
    user_id,
    text,
    task_id,
    task_type,
    next_send, 
    notification_params,
    first_send
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $5
) RETURNING id, created_at, user_id, text, description, task_id, task_type, next_send, done, notification_params, first_send, notify
`

type AddEventParams struct {
	UserID             int32                     `db:"user_id"`
	Text               string                    `db:"text"`
	TaskID             int32                     `db:"task_id"`
	TaskType           TaskType                  `db:"task_type"`
	NextSend           pgtype.Timestamptz        `db:"next_send"`
	NotificationParams domain.NotificationParams `db:"notification_params"`
}

func (q *Queries) AddEvent(ctx context.Context, db DBTX, arg AddEventParams) (Event, error) {
	row := db.QueryRow(ctx, addEvent,
		arg.UserID,
		arg.Text,
		arg.TaskID,
		arg.TaskType,
		arg.NextSend,
		arg.NotificationParams,
	)
	var i Event
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UserID,
		&i.Text,
		&i.Description,
		&i.TaskID,
		&i.TaskType,
		&i.NextSend,
		&i.Done,
		&i.NotificationParams,
		&i.FirstSend,
		&i.Notify,
	)
	return i, err
}

const deleteEvent = `-- name: DeleteEvent :many
DELETE FROM events
WHERE id = $1
RETURNING id, created_at, user_id, text, description, task_id, task_type, next_send, done, notification_params, first_send, notify
`

func (q *Queries) DeleteEvent(ctx context.Context, db DBTX, id int32) ([]Event, error) {
	rows, err := db.Query(ctx, deleteEvent, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Event
	for rows.Next() {
		var i Event
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UserID,
			&i.Text,
			&i.Description,
			&i.TaskID,
			&i.TaskType,
			&i.NextSend,
			&i.Done,
			&i.NotificationParams,
			&i.FirstSend,
			&i.Notify,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getEvent = `-- name: GetEvent :one
SELECT id, created_at, user_id, text, description, task_id, task_type, next_send, done, notification_params, first_send, notify FROM events
WHERE id = $1
`

func (q *Queries) GetEvent(ctx context.Context, db DBTX, id int32) (Event, error) {
	row := db.QueryRow(ctx, getEvent, id)
	var i Event
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UserID,
		&i.Text,
		&i.Description,
		&i.TaskID,
		&i.TaskType,
		&i.NextSend,
		&i.Done,
		&i.NotificationParams,
		&i.FirstSend,
		&i.Notify,
	)
	return i, err
}

const getLatestEvent = `-- name: GetLatestEvent :one
SELECT id, created_at, user_id, text, description, task_id, task_type, next_send, done, notification_params, first_send, notify FROM events
WHERE task_id = $1
  AND task_type = $2
ORDER BY next_send DESC
LIMIT 1
`

type GetLatestEventParams struct {
	TaskID   int32    `db:"task_id"`
	TaskType TaskType `db:"task_type"`
}

func (q *Queries) GetLatestEvent(ctx context.Context, db DBTX, arg GetLatestEventParams) (Event, error) {
	row := db.QueryRow(ctx, getLatestEvent, arg.TaskID, arg.TaskType)
	var i Event
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UserID,
		&i.Text,
		&i.Description,
		&i.TaskID,
		&i.TaskType,
		&i.NextSend,
		&i.Done,
		&i.NotificationParams,
		&i.FirstSend,
		&i.Notify,
	)
	return i, err
}

const getNearestEventTime = `-- name: GetNearestEventTime :one
SELECT next_send FROM events
WHERE done = false
  AND notify = true 
ORDER BY next_send ASC
LIMIT 1
`

func (q *Queries) GetNearestEventTime(ctx context.Context, db DBTX) (pgtype.Timestamptz, error) {
	row := db.QueryRow(ctx, getNearestEventTime)
	var next_send pgtype.Timestamptz
	err := row.Scan(&next_send)
	return next_send, err
}

const listNotDoneEvents = `-- name: ListNotDoneEvents :many
SELECT id, created_at, user_id, text, description, task_id, task_type, next_send, done, notification_params, first_send, notify FROM events
WHERE user_id = $1
  AND done = false
  AND next_send < NOW()
ORDER BY next_send ASC
`

func (q *Queries) ListNotDoneEvents(ctx context.Context, db DBTX, userID int32) ([]Event, error) {
	rows, err := db.Query(ctx, listNotDoneEvents, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Event
	for rows.Next() {
		var i Event
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UserID,
			&i.Text,
			&i.Description,
			&i.TaskID,
			&i.TaskType,
			&i.NextSend,
			&i.Done,
			&i.NotificationParams,
			&i.FirstSend,
			&i.Notify,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listNotSendedEvents = `-- name: ListNotSendedEvents :many
SELECT id, created_at, user_id, text, description, task_id, task_type, next_send, done, notification_params, first_send, notify FROM events
WHERE next_send <= $1
  AND done = false
  AND notify = true
`

func (q *Queries) ListNotSendedEvents(ctx context.Context, db DBTX, till pgtype.Timestamptz) ([]Event, error) {
	rows, err := db.Query(ctx, listNotSendedEvents, till)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Event
	for rows.Next() {
		var i Event
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UserID,
			&i.Text,
			&i.Description,
			&i.TaskID,
			&i.TaskType,
			&i.NextSend,
			&i.Done,
			&i.NotificationParams,
			&i.FirstSend,
			&i.Notify,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listUserDailyEvents = `-- name: ListUserDailyEvents :many
SELECT id, created_at, user_id, text, description, task_id, task_type, next_send, done, notification_params, first_send, notify FROM events
WHERE user_id = $1
  AND next_send + $2  BETWEEN CURRENT_DATE AND CURRENT_DATE + 1
ORDER BY next_send ASC
`

type ListUserDailyEventsParams struct {
	UserID     int32              `db:"user_id"`
	TimeOffset pgtype.Timestamptz `db:"time_offset"`
}

func (q *Queries) ListUserDailyEvents(ctx context.Context, db DBTX, arg ListUserDailyEventsParams) ([]Event, error) {
	rows, err := db.Query(ctx, listUserDailyEvents, arg.UserID, arg.TimeOffset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Event
	for rows.Next() {
		var i Event
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UserID,
			&i.Text,
			&i.Description,
			&i.TaskID,
			&i.TaskType,
			&i.NextSend,
			&i.Done,
			&i.NotificationParams,
			&i.FirstSend,
			&i.Notify,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listUserEvents = `-- name: ListUserEvents :many
SELECT DISTINCT e.id, e.created_at, e.user_id, e.text, e.description, e.task_id, e.task_type, e.next_send, e.done, e.notification_params, e.first_send, e.notify 
FROM events as e
LEFT JOIN smth2tags as s2t
  ON e.id = s2t.smth_id
LEFT JOIN tags as t
  ON s2t.tag_id = t.id
WHERE e.user_id = $1
  AND next_send BETWEEN $2 AND $3
  AND (
    t.id = ANY ($4::int[]) 
    OR array_length($4::int[], 1) is null
  )
ORDER BY next_send DESC
LIMIT $6 OFFSET $5
`

type ListUserEventsParams struct {
	UserID   int32              `db:"user_id"`
	FromTime pgtype.Timestamptz `db:"from_time"`
	ToTime   pgtype.Timestamptz `db:"to_time"`
	TagIds   []int32            `db:"tag_ids"`
	Off      int32              `db:"off"`
	Lim      int32              `db:"lim"`
}

type ListUserEventsRow struct {
	Event Event `db:"event"`
}

func (q *Queries) ListUserEvents(ctx context.Context, db DBTX, arg ListUserEventsParams) ([]ListUserEventsRow, error) {
	rows, err := db.Query(ctx, listUserEvents,
		arg.UserID,
		arg.FromTime,
		arg.ToTime,
		arg.TagIds,
		arg.Off,
		arg.Lim,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ListUserEventsRow
	for rows.Next() {
		var i ListUserEventsRow
		if err := rows.Scan(
			&i.Event.ID,
			&i.Event.CreatedAt,
			&i.Event.UserID,
			&i.Event.Text,
			&i.Event.Description,
			&i.Event.TaskID,
			&i.Event.TaskType,
			&i.Event.NextSend,
			&i.Event.Done,
			&i.Event.NotificationParams,
			&i.Event.FirstSend,
			&i.Event.Notify,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateEvent = `-- name: UpdateEvent :one
UPDATE events
SET text = $1,
    next_send = $2,
    first_send = $3,
    done = $4
WHERE id = $5
RETURNING id, created_at, user_id, text, description, task_id, task_type, next_send, done, notification_params, first_send, notify
`

type UpdateEventParams struct {
	Text      string             `db:"text"`
	NextSend  pgtype.Timestamptz `db:"next_send"`
	FirstSend pgtype.Timestamptz `db:"first_send"`
	Done      bool               `db:"done"`
	ID        int32              `db:"id"`
}

func (q *Queries) UpdateEvent(ctx context.Context, db DBTX, arg UpdateEventParams) (Event, error) {
	row := db.QueryRow(ctx, updateEvent,
		arg.Text,
		arg.NextSend,
		arg.FirstSend,
		arg.Done,
		arg.ID,
	)
	var i Event
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UserID,
		&i.Text,
		&i.Description,
		&i.TaskID,
		&i.TaskType,
		&i.NextSend,
		&i.Done,
		&i.NotificationParams,
		&i.FirstSend,
		&i.Notify,
	)
	return i, err
}
