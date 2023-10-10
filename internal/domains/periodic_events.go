package domains

import (
	"math/rand"
	"time"
)

const timeDay = 24 * time.Hour

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

func (pe PeriodicEvent) NextNotification() PeriodicEventNotification {
	minDays := int(pe.SmallestPeriod / timeDay)
	maxDays := int(pe.BiggestPeriod / timeDay)
	days := minDays + rand.Intn(maxDays-minDays) //nolint:gosec // no need to use crypto rand
	dayBeginning := time.Now().Add(time.Duration(days) * timeDay).Truncate(timeDay)
	sendTime := dayBeginning.Add(pe.Start)

	return PeriodicEventNotification{
		ID:              0,
		PeriodicEventID: pe.ID,
		SendTime:        sendTime,
		Sended:          false,
		Done:            false,
	}
}

type PeriodicEventNotification struct {
	ID              int
	PeriodicEventID int
	SendTime        time.Time
	Sended          bool
	Done            bool
}

type PeriodicEventWithNotification struct {
	Event        PeriodicEvent
	Notification PeriodicEventNotification
}
