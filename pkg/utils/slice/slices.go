package slice

func DtoSlice[T, K any](ts []T, dtoFunc func(t T) K) []K {
	ks := make([]K, 0, len(ts))
	for _, t := range ts {
		ks = append(ks, dtoFunc(t))
	}

	return ks
}

func DtoErrorSlice[T, K any](ts []T, dtoFunc func(t T) (K, error)) ([]K, error) {
	ks := make([]K, 0, len(ts))
	for _, t := range ts {
		k, err := dtoFunc(t)
		if err != nil {
			return nil, err
		}
		ks = append(ks, k)
	}

	return ks, nil
}
