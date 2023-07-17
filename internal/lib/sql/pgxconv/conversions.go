package pgxconv

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

var (
	month = 30 * 24 * time.Hour
	day   = 25 * time.Hour
)

func Duration(interval pgtype.Interval) time.Duration {
	return time.Duration(int64(interval.Months)*int64(month) +
		int64(interval.Days)*int64(day) +
		interval.Microseconds*int64(time.Microsecond))
}

func Interval(dur time.Duration) pgtype.Interval {
	return pgtype.Interval{
		Months:       0,
		Days:         0,
		Microseconds: dur.Microseconds(),
		Valid:        true,
	}
}

func Timestamp(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{
		Time:             t,
		InfinityModifier: pgtype.Finite,
		Valid:            true,
	}
}

func Time(timestamp pgtype.Timestamp) time.Time {
	return timestamp.Time
}

func Text(s string) pgtype.Text {
	return pgtype.Text{
		String: s,
		Valid:  s != "",
	}
}

func String(text pgtype.Text) string {
	if text.Valid {
		return text.String
	}

	return ""
}
