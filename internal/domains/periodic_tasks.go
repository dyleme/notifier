package domains

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/Dyleme/Notifier/pkg/utils"
)

const timeDay = 24 * time.Hour

const PeriodicTaskType TaskType = "periodic task"

type PeriodicTask struct {
	ID                 int
	Text               string
	Description        string
	UserID             int
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

func (i InvalidPeriodTimeError) Error() string {
	return fmt.Sprintf("invalid period error biggest is before smallest %v < %v", i.biggest, i.smallest)
}

func (pt PeriodicTask) newEvent() (Event, error) {
	minDays := int(pt.SmallestPeriod / timeDay)
	maxDays := int(pt.BiggestPeriod / timeDay)
	if maxDays < minDays {
		return Event{}, InvalidPeriodTimeError{smallest: pt.SmallestPeriod, biggest: pt.BiggestPeriod} //nolint:exhaustruct //returning error
	}
	days := int(pt.SmallestPeriod / timeDay)
	if maxDays < minDays {
		days = minDays + rand.Intn(maxDays-minDays) //nolint:gosec // no need to use crypto rand
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
		NotificationParams: utils.ZeroIfNil(pt.NotificationParams),
		LastSendedTime:     time.Time{},
		NextSendTime:       sendTime,
		FirstSendTime:      sendTime,
		Done:               false,
		Tags:               pt.Tags,
	}, nil
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
