package domains

import (
	"time"
)

type Event struct {
	ID                 int
	UserID             int
	Text               string
	Description        string
	Start              time.Time
	Done               bool
	NotificationParams *NotificationParams
	SendTime           time.Time
	Sended             bool
}

func (e *Event) IsGettingDone(newDone bool) bool {
	if !e.Done && newDone {
		return true
	}

	return false
}

func EventFromTask(t Task, start time.Time, description string) Event {
	return Event{ //nolint:exhaustruct  // TODO: We dont know Service id
		UserID:             t.UserID,
		Text:               t.Text,
		Start:              start,
		Done:               false,
		Description:        description,
		Sended:             false,
		SendTime:           start,
		NotificationParams: nil,
	}
}
