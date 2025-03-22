package domain

import (
	"fmt"
	"time"

	"github.com/Dyleme/Notifier/pkg/utils"
)

type EventCreator interface {
	newEvent(now time.Time) (Event, error)
}

func CreateEvent(eventCreator EventCreator, defaultParams NotificationParams) (Event, error) {
	event, err := eventCreator.newEvent(time.Now())
	if err != nil {
		return Event{}, fmt.Errorf("new event: %w", err)
	}

	if event.Notify {
		if utils.IsZero(event.NotificationParams) {
			event.NotificationParams = defaultParams
		}
	}

	if err := event.Validate(); err != nil {
		return Event{}, err
	}

	return event, nil
}
