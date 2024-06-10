package utils

func Ptr[T any](t T) *T {
	return &t
}

func ZeroIfNil[T comparable](t *T) T {
	if t == nil {
		var zero T

		return zero
	}

	return *t
}
