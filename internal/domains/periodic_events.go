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
	Notification       PeriodicEventNotification
}

type InvalidPeriodTimeError struct {
	smallest time.Duration
	biggest  time.Duration
}

func (i InvalidPeriodTimeError) Error() string {
	return fmt.Sprintf("invalid period error biggest is before smallest %v < %v", i.biggest, i.smallest)
}

func (pe PeriodicEvent) NextNotification() (PeriodicEventNotification, error) {
	minDays := int(pe.SmallestPeriod / timeDay)
	maxDays := int(pe.BiggestPeriod / timeDay)
	if maxDays < minDays {
		return PeriodicEventNotification{}, InvalidPeriodTimeError{smallest: pe.SmallestPeriod, biggest: pe.BiggestPeriod} //nolint:exhaustruct //returning error
	}
	days := int(pe.SmallestPeriod / timeDay)
	if maxDays < minDays {
		days = minDays + rand.Intn(maxDays-minDays) //nolint:gosec // no need to use crypto rand
	}
	dayBeginning := time.Now().Add(time.Duration(days) * timeDay).Truncate(timeDay)
	sendTime := dayBeginning.Add(pe.Start)

	return PeriodicEventNotification{
		ID:              0,
		PeriodicEventID: pe.ID,
		SendTime:        sendTime,
		Sended:          false,
		Done:            false,
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

type PeriodicEventNotification struct {
	ID              int
	PeriodicEventID int
	SendTime        time.Time
	Sended          bool
	Done            bool
}
