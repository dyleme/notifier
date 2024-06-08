package domains

import (
	"fmt"
	"math/rand"
	"time"
)

const timeDay = 24 * time.Hour

const PeriodicEventType EventType = "periodic event"

type PeriodicEvent struct {
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

func (pe PeriodicEvent) NewNotification(t time.Time) (Notification, error) {
	minDays := int(pe.SmallestPeriod / timeDay)
	maxDays := int(pe.BiggestPeriod / timeDay)
	if maxDays < minDays {
		return Notification{}, InvalidPeriodTimeError{smallest: pe.SmallestPeriod, biggest: pe.BiggestPeriod} //nolint:exhaustruct //returning error
	}
	days := int(pe.SmallestPeriod / timeDay)
	if maxDays < minDays {
		days = minDays + rand.Intn(maxDays-minDays) //nolint:gosec // no need to use crypto rand
	}
	dayBeginning := t.Add(time.Duration(days) * timeDay).Truncate(timeDay)
	sendTime := dayBeginning.Add(pe.Start)

	return Notification{
		ID:          0,
		UserID:      pe.UserID,
		Text:        pe.Text,
		Description: pe.Description,
		EventType:   PeriodicEventType,
		EventID:     pe.ID,
		Params:      pe.NotificationParams,
		SendTime:    sendTime,
		Sended:      false,
		Done:        false,
	}, nil
}

func (pe PeriodicEvent) NeedRegenerateNotification(updated PeriodicEvent) bool {
	if updated.Start != pe.Start {
		return true
	}

	if updated.BiggestPeriod != pe.BiggestPeriod || updated.SmallestPeriod != pe.SmallestPeriod {
		return true
	}

	return false
}

func (pe PeriodicEvent) BelongsTo(userID int) bool {
	return pe.UserID == userID
}

type PeriodicEventNotification struct {
	ID              int
	PeriodicEventID int
	SendTime        time.Time
	Sended          bool
	Done            bool
}
