package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/Dyleme/Notifier/internal/domain/apperr"
)

const timeDay = 24 * time.Hour

type InvalidPeriodTimeError struct {
	smallest time.Duration
	biggest  time.Duration
}

var ErrNotificaitonParamsRequired = errors.New("notification params is required")

func (i InvalidPeriodTimeError) Error() string {
	return fmt.Sprintf("invalid period error biggest is before smallest %v < %v", i.biggest, i.smallest)
}

func (pt PeriodicTask) Validate() error {
	minDays := int(pt.SmallestPeriod / timeDay)
	maxDays := int(pt.BiggestPeriod / timeDay)
	if maxDays < minDays {
		return InvalidPeriodTimeError{smallest: pt.SmallestPeriod, biggest: pt.BiggestPeriod}
	}

	return nil
}

func (pt PeriodicTask) BelongsTo(userID int) error {
	if pt.UserID == userID {
		return nil
	}

	return apperr.NewNotBelongToUserError("periodic task", pt.ID, pt.UserID, userID)
}

func (pt PeriodicTask) TimeParamsHasChanged(updT PeriodicTask) bool {
	return pt.SmallestPeriod != updT.SmallestPeriod ||
		pt.BiggestPeriod != updT.BiggestPeriod ||
		pt.Start != updT.Start
}
