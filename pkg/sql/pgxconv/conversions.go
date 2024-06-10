package pgxconv

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

var (
	month = 30 * 24 * time.Hour
	day   = 24 * time.Hour
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

var (
	minTime = time.Unix(0, 0)
	maxTime = time.Unix(1<<63-1, 0)
)

func Time(timestamp pgtype.Timestamp) time.Time {
	return timestamp.Time
}

func Timestamptz(t time.Time) pgtype.Timestamptz {
	if t == minTime {
		return pgtype.Timestamptz{
			Time:             time.Time{},
			InfinityModifier: pgtype.NegativeInfinity,
			Valid:            true,
		}
	}
	if t == maxTime {
		return pgtype.Timestamptz{
			Time:             time.Time{},
			InfinityModifier: pgtype.Infinity,
			Valid:            true,
		}
	}

	return pgtype.Timestamptz{
		Time:             t,
		InfinityModifier: pgtype.Finite,
		Valid:            true,
	}
}

func TimeWithZone(timestamps pgtype.Timestamptz) time.Time {
	return timestamps.Time
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

func ByteSlice(text pgtype.Text) []byte {
	if text.Valid {
		return []byte(text.String)
	}

	return nil
}

func Int4(i *int) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{Int32: 0, Valid: false}
	}

	return pgtype.Int4{Int32: int32(*i), Valid: true}
}

func Int(i pgtype.Int4) int {
	if i.Valid {
		return int(i.Int32)
	}

	return 0
}

func PgBool(b bool) pgtype.Bool {
	return pgtype.Bool{
		Bool:  b,
		Valid: true,
	}
}
