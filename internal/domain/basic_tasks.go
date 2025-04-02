package domain

import (
	"fmt"
	"time"

	"github.com/Dyleme/Notifier/internal/domain/apperr"
	utils "github.com/Dyleme/Notifier/pkg/utils/ptr"
)

const BasicTaskType TaskType = "basic task"

type BasicTask struct {
	ID                 int
	UserID             int
	Text               string
	Description        string
	Notify             bool
	Start              time.Time
	NotificationParams NotificationParams
	Tags               []Tag
}

func (bt BasicTask) newEvent(_ time.Time, defaultNotifParams NotificationParams) (Event, error) { //nolint:unparam //need for interface impolementation
	if bt.Notify && utils.IsZero(bt.NotificationParams) {
		bt.NotificationParams = defaultNotifParams
	}

	return Event{
		ID:                 0,
		UserID:             bt.UserID,
		Text:               bt.Text,
		Description:        bt.Description,
		TaskType:           BasicTaskType,
		TaskID:             bt.ID,
		NotificationParams: bt.NotificationParams,
		NextSend:           bt.Start,
		FirstSend:          bt.Start,
		Done:               false,
		Tags:               bt.Tags,
		Notify:             bt.Notify,
	}, nil
}

func (bt BasicTask) UpdatedEvent(ev Event) (Event, error) {
	updatedEvent, err := bt.newEvent(time.Now(), ev.NotificationParams)
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

	return apperr.NewNotBelongToUserError("basic task", bt.ID, bt.UserID, userID)
}
