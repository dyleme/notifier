package domains

import (
	"time"
)

type Event struct {
	ID           int
	UserID       int
	Text         string
	Description  string
	Start        time.Time
	Done         bool
	Notification Notification
}

type Notification struct {
	Sended             bool                `json:"sended"`
	NotificationParams *NotificationParams `json:"notification_params"`
}

func (e *Event) IsGettingDone(newDone bool) bool {
	if !e.Done && newDone {
		return true
	}

	return false
}

func EventFromTask(t Task, start time.Time, description string) Event {
	return Event{ //nolint:exhaustruct  // TODO: We dont know Event id
		UserID:      t.UserID,
		Text:        t.Text,
		Start:       start,
		Done:        false,
		Description: description,
	}
}
