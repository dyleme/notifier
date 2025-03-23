package domain

import "github.com/Dyleme/Notifier/internal/domain/apperr"

type User struct {
	ID             int
	TgNickname     string
	PasswordHash   []byte
	TGID           int
	TimeZoneOffset int
	IsTimeZoneDST  bool
}

type TimeZoneOffset int

func (to TimeZoneOffset) Valid() error {
	if -24 < to && to < 24 {
		return apperr.InvalidOffsetError{Offset: int(to)}
	}

	return nil
}
