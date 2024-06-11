package domains

import (
	"time"
)

const BasicTaskType TaskType = "basic task"

type BasicTask struct {
	ID                 int
	UserID             int
	Text               string
	Description        string
	Start              time.Time
	NotificationParams *NotificationParams
}

func (bt BasicTask) NewEvent() Event {
	return Event{
		ID:          0,
		UserID:      bt.UserID,
		Text:        bt.Text,
		Description: bt.Description,
		TaskType:    BasicTaskType,
		TaskID:      bt.ID,
		Params:      bt.NotificationParams,
		SendTime:    bt.Start,
		Sended:      false,
		Done:        false,
	}
}

func (bt BasicTask) BelongsTo(userID int) bool {
	return bt.UserID == userID
}
