package domains

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"time"
)

const timeDay = 24 * time.Hour

const PeriodicTaskType TaskType = "periodic task"

type PeriodicTask struct {
	ID                 int
	Text               string
	Description        string
	UserID             int
	Notify             bool
	Start              time.Duration // Event time from beginning of day
	SmallestPeriod     time.Duration
	BiggestPeriod      time.Duration
	NotificationParams *NotificationParams
	Tags               []Tag
}

type InvalidPeriodTimeError struct {
	smallest time.Duration
	biggest  time.Duration
}

var ErrNotificaitonParamsRequired = errors.New("notification params is required")

func (i InvalidPeriodTimeError) Error() string {
	return fmt.Sprintf("invalid period error biggest is before smallest %v < %v", i.biggest, i.smallest)
}

func (pt PeriodicTask) newEvent() (Event, error) {
	err := pt.Validate()
	if err != nil {
		return Event{}, err
	}
	minDays := int(pt.SmallestPeriod / timeDay)
	maxDays := int(pt.BiggestPeriod / timeDay)
	days := minDays
	if diff := maxDays - minDays; diff > 0 { // need if as rand.IntN panics if diff == 0
		days = minDays + rand.IntN(diff) //nolint:gosec // no need to use crypto rand
	}
	dayBeginning := time.Now().Add(time.Duration(days) * timeDay).Truncate(timeDay)
	sendTime := dayBeginning.Add(pt.Start)

	return Event{
		ID:                 0,
		UserID:             pt.UserID,
		Text:               pt.Text,
		Description:        pt.Description,
		TaskType:           PeriodicTaskType,
		TaskID:             pt.ID,
		NotificationParams: pt.NotificationParams,
		LastSendedTime:     time.Time{},
		Time:               sendTime,
		FirstTime:          sendTime,
		Done:               false,
		Tags:               pt.Tags,
		Notify:             pt.Notify,
	}, nil
}

func (pt PeriodicTask) Validate() error {
	minDays := int(pt.SmallestPeriod / timeDay)
	maxDays := int(pt.BiggestPeriod / timeDay)
	if maxDays < minDays {
		return InvalidPeriodTimeError{smallest: pt.SmallestPeriod, biggest: pt.BiggestPeriod} //nolint:exhaustruct //returning error
	}

	if pt.Notify && pt.NotificationParams == nil {
		return ErrNotificaitonParamsRequired
	}

	return nil
}

func (pt PeriodicTask) BelongsTo(userID int) error {
	if pt.UserID == userID {
		return nil
	}

	return NewNotBelongToUserError("periodic task", pt.ID, pt.UserID, userID)
}

func (pt PeriodicTask) TimeParamsHasChanged(updT PeriodicTask) bool {
	return pt.SmallestPeriod != updT.SmallestPeriod ||
		pt.BiggestPeriod != updT.BiggestPeriod ||
		pt.Start != updT.Start
}
