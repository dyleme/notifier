package compare

import (
	"time"
)

func TimeCmpWithoutZero(a, b time.Time) int {
	if a.Equal(b) {
		return 0
	}
	if a.IsZero() {
		return 1
	}
	if b.IsZero() || a.Before(b) {
		return -1
	}

	return 1
}

func Zero[T comparable](t T) bool {
	var zero T

	return t == zero
}
