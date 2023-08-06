package ptr

func Ptr[T any](t T) *T {
	return &t
}
