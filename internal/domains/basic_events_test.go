package domains_test

import (
	"testing"
	"time"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/stretchr/testify/require"
)

func TestBasicEvent_NewNotification(t *testing.T) {
	t.Parallel()
	t.Run("check mapping", func(t *testing.T) {
		t.Parallel()
		basicEvent := domains.BasicEvent{
			ID:          1,
			UserID:      2,
			Text:        "text",
			Description: "description",
			Start:       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			NotificationParams: &domains.NotificationParams{
				Period: time.Hour,
				Params: domains.Params{
					Telegram: 3,
				},
			},
		}
		actual := basicEvent.NewNotification()

		expected := domains.Notification{
			ID:          0,
			UserID:      2,
			Text:        "text",
			Description: "description",
			EventType:   domains.BasicEventType,
			EventID:     1,
			Params:      basicEvent.NotificationParams,
			SendTime:    time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			Sended:      false,
			Done:        false,
		}

		require.Equal(t, expected, actual)
	})
}
