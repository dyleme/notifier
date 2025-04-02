package domain

import (
	"fmt"
	"time"
)

type EventCreator interface {
	newEvent(now time.Time, deafultNotification NotificationParams) (Event, error)
}

func CreateEvent(eventCreator EventCreator, defaultParams NotificationParams) (Event, error) {
	event, err := eventCreator.newEvent(time.Now(), defaultParams)
	if err != nil {
		return Event{}, fmt.Errorf("new event: %w", err)
	}

	if err := event.Validate(); err != nil {
		return Event{}, err
	}

	return event, nil
}
