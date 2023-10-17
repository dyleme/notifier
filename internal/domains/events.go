package domains

import (
	"time"
)

const BasicEventType EventType = "basic event"

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
