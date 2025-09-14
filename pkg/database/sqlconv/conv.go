package sqlconv

import "database/sql"

func ToBool(i int64) bool {
	return i != 0
}

func BoolToInt(b bool) int64 {
	if b {
		return 1
	}

	return 0
}

func Nullable[T comparable](t T) sql.Null[T] {
	var zero T
	if t == zero {
		return sql.Null[T]{Valid: false}
	}

	return sql.Null[T]{V: t, Valid: true}
}

func NullableString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}

	return sql.NullString{String: s, Valid: true}
}
