package domain

import (
	"time"

	"github.com/Dyleme/Notifier/internal/domain/apperr"
)

type User struct {
	ID                        int
	TGID                      int
	TimeZoneOffset            int
	IsTimeZoneDST             bool
	DefaultNotificationPeriod time.Duration
}

func (u User) Location() *time.Location {
	return time.FixedZone("Temporary", u.TimeZoneOffset*int(time.Hour/time.Second))
}

type TimeZoneOffset int

func (to TimeZoneOffset) Valid() error {
	if to < -24 || to > 24 {
		return apperr.InvalidOffsetError{Offset: int(to)}
	}

	return nil
}
