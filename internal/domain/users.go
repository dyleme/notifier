package domain

import "github.com/Dyleme/Notifier/internal/domain/apperr"

type User struct {
	ID             int
	TGID           int
	TimeZoneOffset int
	IsTimeZoneDST  bool
}

type TimeZoneOffset int

func (to TimeZoneOffset) Valid() error {
	if to < -24 || to > 24 {
		return apperr.InvalidOffsetError{Offset: int(to)}
	}

	return nil
}
