package ptr

func On[T any](t T) *T {
	return &t
}

func ZeroIfNil[T any](t *T) T {
	if t == nil {
		var zero T

		return zero
	}

	return *t
}

func NilIfZero[T comparable](t T) *T {
	var zero T
	if t == zero {
		return nil
	}

	return &t
}

func IsZero[T comparable](t T) bool {
	var zero T

	return t == zero
}
