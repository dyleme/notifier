package domains

import (
	"fmt"
	"time"

	"github.com/Dyleme/Notifier/pkg/utils"
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

func (bt BasicTask) newEvent() (Event, error) { //nolint:unparam //need for interface impolementation
	return Event{
		ID:                 0,
		UserID:             bt.UserID,
		Text:               bt.Text,
		Description:        bt.Description,
		TaskType:           BasicTaskType,
		TaskID:             bt.ID,
		NotificationParams: utils.ZeroIfNil(bt.NotificationParams),
		LastSendedTime:     time.Time{},
		NextSendTime:       bt.Start,
		FirstSendTime:      bt.Start,
		Done:               false,
	}, nil
}

func (bt BasicTask) UpdatedEvent(ev Event) (Event, error) {
	updatedEvent, err := bt.newEvent()
	if err != nil {
		return Event{}, fmt.Errorf("new event: %w", err)
	}

	updatedEvent.ID = ev.ID
	updatedEvent.NotificationParams = ev.NotificationParams

	return updatedEvent, nil
}

func (bt BasicTask) BelongsTo(userID int) error {
	if bt.UserID == userID {
		return nil
	}

	return NewNotBelongToUserError("basic task", bt.ID, bt.UserID, userID)
}
