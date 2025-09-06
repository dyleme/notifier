package sqliteconv

func ToBool(i int64) bool {
	return i != 0
}

func BoolToInt(b bool) int64 {
	if b {
		return 1
	}

	return 0
}
