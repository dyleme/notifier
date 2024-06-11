package domains

import (
	"fmt"
	"math/rand"
	"time"
)

const timeDay = 24 * time.Hour

const PeriodicTaskType TaskType = "periodic task"

type PeriodicTask struct {
	ID                 int
	Text               string
	Description        string
	UserID             int
	Start              time.Duration // Notification time from beginning of day
	SmallestPeriod     time.Duration
	BiggestPeriod      time.Duration
	NotificationParams *NotificationParams
}

type InvalidPeriodTimeError struct {
	smallest time.Duration
	biggest  time.Duration
}

func (i InvalidPeriodTimeError) Error() string {
	return fmt.Sprintf("invalid period error biggest is before smallest %v < %v", i.biggest, i.smallest)
}

func (pt PeriodicTask) NewNotification(t time.Time) (Notification, error) {
	minDays := int(pt.SmallestPeriod / timeDay)
	maxDays := int(pt.BiggestPeriod / timeDay)
	if maxDays < minDays {
		return Notification{}, InvalidPeriodTimeError{smallest: pt.SmallestPeriod, biggest: pt.BiggestPeriod} //nolint:exhaustruct //returning error
	}
	days := int(pt.SmallestPeriod / timeDay)
	if maxDays < minDays {
		days = minDays + rand.Intn(maxDays-minDays) //nolint:gosec // no need to use crypto rand
	}
	dayBeginning := t.Add(time.Duration(days) * timeDay).Truncate(timeDay)
	sendTime := dayBeginning.Add(pt.Start)

	return Notification{
		ID:          0,
		UserID:      pt.UserID,
		Text:        pt.Text,
		Description: pt.Description,
		TaskType:    PeriodicTaskType,
		TaskID:      pt.ID,
		Params:      pt.NotificationParams,
		SendTime:    sendTime,
		Sended:      false,
		Done:        false,
	}, nil
}

func (pt PeriodicTask) NeedRegenerateNotification(updated PeriodicTask) bool {
	if updated.Start != pt.Start {
		return true
	}

	if updated.BiggestPeriod != pt.BiggestPeriod || updated.SmallestPeriod != pt.SmallestPeriod {
		return true
	}

	return false
}

func (pt PeriodicTask) BelongsTo(userID int) bool {
	return pt.UserID == userID
}

type PeriodicTaskNotification struct {
	ID             int
	PeriodicTaskID int
	SendTime       time.Time
	Sended         bool
	Done           bool
}
