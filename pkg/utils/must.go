package utils

func Must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}

	return t
}

func NoErr(err error) {
	if err != nil {
		panic(err)
	}
}

func NoErrFunc(_ any, err error) {
	if err != nil {
		panic(err)
	}
}
