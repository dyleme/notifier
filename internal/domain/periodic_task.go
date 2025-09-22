package domain

import (
	"errors"
	"fmt"
	"math/rand/v2"
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

const (
	smallestPeriodKey CreationParamKey = "smallest_period"
	biggestPeriodKey  CreationParamKey = "biggest_period"
)

type PeriodicTask struct {
	Task
}

func (pt PeriodicTask) SmallestPeriod() time.Duration {
	f := pt.EventCreationParams[smallestPeriodKey].(float64) //nolint:errcheck,forcetypeassert //hope nothing will broke

	return time.Duration(f)
}

func (pt PeriodicTask) BiggestPeriod() time.Duration {
	f := pt.EventCreationParams[biggestPeriodKey].(float64) //nolint:errcheck,forcetypeassert //hope nothing will broke

	return time.Duration(f)
}

func NewPeriodicTask(params TaskCreationParams, smallestPeriod, biggestPeriod time.Duration) PeriodicTask {
	return PeriodicTask{
		Task: Task{
			ID:          params.ID,
			Text:        params.Text,
			Description: params.Description,
			UserID:      params.UserID,
			Type:        Periodic,
			Start:       params.Start,
			EventCreationParams: map[CreationParamKey]any{
				smallestPeriodKey: float64(smallestPeriod),
				biggestPeriodKey:  float64(biggestPeriod),
			},
		},
	}
}

func (pt PeriodicTask) NewEvent(now time.Time) Sending {
	minDays := int(pt.SmallestPeriod() / timeDay)
	maxDays := int(pt.BiggestPeriod() / timeDay)
	days := minDays
	if diff := maxDays - minDays; diff > 0 {
		days += rand.IntN(diff) //nolint:gosec // no need for security
	}
	dayBeginning := now.Add(time.Duration(days) * timeDay).Truncate(timeDay)
	sendTime := dayBeginning.Add(pt.Start)

	return Sending{
		TaskID:          pt.ID,
		Done:            false,
		OriginalSending: sendTime,
		NextSending:     sendTime,
	}
}

func (pt PeriodicTask) Validate() error {
	minDays := int(pt.SmallestPeriod() / timeDay)
	maxDays := int(pt.BiggestPeriod() / timeDay)
	if maxDays < minDays {
		return InvalidPeriodTimeError{smallest: pt.SmallestPeriod(), biggest: pt.BiggestPeriod()}
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
	return pt.SmallestPeriod() != updT.SmallestPeriod() ||
		pt.BiggestPeriod() != updT.BiggestPeriod() ||
		pt.Start != updT.Start
}
