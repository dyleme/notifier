package domains_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Dyleme/Notifier/internal/domains"
)

func TestNewSendingEvent(t *testing.T) {
	t.Parallel()
	t.Run("check mapping", func(t *testing.T) {
		t.Parallel()
		params := domains.NotificationParams{
			Period: time.Minute,
			Params: domains.Params{
				Telegram: 5,
			},
		}
		event := domains.Event{
			ID:                 1,
			UserID:             2,
			Text:               "text",
			Description:        "description",
			TaskType:           domains.BasicTaskType,
			TaskID:             3,
			NotificationParams: params,
			SendTime:           time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			Sended:             true,
			Done:               false,
		}
		actual := domains.NewSendingEvent(event)

		expected := domains.SendingEvent{
			EventID:     1,
			UserID:      2,
			Message:     "text",
			Description: "description",
			Params:      params,
			SendTime:    time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		}

		require.Equal(t, expected, actual)
	})
}
