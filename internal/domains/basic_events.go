package domains

import (
	"time"
)

const BasicEventType EventType = "basic event"

type BasicEvent struct {
	ID                 int
	UserID             int
	Text               string
	Description        string
	Start              time.Time
	NotificationParams *NotificationParams
}

func (be BasicEvent) NewNotification() Notification {
	return Notification{
		ID:          0,
		UserID:      be.UserID,
		Text:        be.Text,
		Description: be.Description,
		EventType:   BasicEventType,
		EventID:     be.ID,
		Params:      be.NotificationParams,
		SendTime:    be.Start,
		Sended:      false,
		Done:        false,
	}
}

func (be BasicEvent) BelongsTo(userID int) bool {
	return be.UserID == userID
}

func BasicEventFromTask(t Task, start time.Time, description string) BasicEvent {
	return BasicEvent{ //nolint:exhaustruct  // TODO: We dont know Service id
		UserID:             t.UserID,
		Text:               t.Text,
		Start:              start,
		Description:        description,
		NotificationParams: nil,
	}
}
