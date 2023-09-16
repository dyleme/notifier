package sqlnull

import (
	"database/sql"
	"time"
)

func ToSQLTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Time: time.Time{}, Valid: false}
	}

	return sql.NullTime{Valid: true, Time: *t}
}

func ToPtrTime(sqlTime sql.NullTime) *time.Time {
	if sqlTime.Valid {
		return &sqlTime.Time
	}

	return nil
}
