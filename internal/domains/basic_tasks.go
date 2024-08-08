package domains

import (
	"fmt"
	"time"
)

const BasicTaskType TaskType = "basic task"

type BasicTask struct {
	ID                 int
	UserID             int
	Text               string
	Description        string
	Notify             bool
	Start              time.Time
	NotificationParams *NotificationParams
	Tags               []Tag
}

func (bt BasicTask) newEvent() (Event, error) { //nolint:unparam //need for interface impolementation
	return Event{
		ID:                 0,
		UserID:             bt.UserID,
		Text:               bt.Text,
		Description:        bt.Description,
		TaskType:           BasicTaskType,
		TaskID:             bt.ID,
		NotificationParams: bt.NotificationParams,
		LastSendedTime:     time.Time{},
		Time:               bt.Start,
		FirstTime:          bt.Start,
		Done:               false,
		Tags:               bt.Tags,
		Notify:             bt.Notify,
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
