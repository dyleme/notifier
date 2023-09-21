package models

type User struct {
	ID             int
	Email          string
	PasswordHash   []byte
	TGID           int
	TimeZoneOffset int
	IsTimeZoneDST  bool
}

type TimeZoneOffset int

func (to TimeZoneOffset) IsValid() bool {
	if -24 < to && to < 24 {
		return true
	}

	return false
}
