package domains

import (
	"fmt"

	"github.com/Dyleme/Notifier/pkg/utils"
)

type EventCreator interface {
	newEvent() (Event, error)
}

func CreateEvent(eventCreator EventCreator, defaultParams NotificationParams) (Event, error) {
	event, err := eventCreator.newEvent()
	if err != nil {
		return Event{}, fmt.Errorf("failed to create event: %w", err)
	}

	if utils.IsZero(event.NotificationParams) {
		event.NotificationParams = defaultParams
	}

	return event, nil
}
