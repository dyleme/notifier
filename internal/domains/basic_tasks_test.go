package domains

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBasicTask_newEvent(t *testing.T) {
	t.Parallel()
	t.Run("check mapping", func(t *testing.T) {
		t.Parallel()
		basicTask := BasicTask{
			ID:          1,
			UserID:      2,
			Text:        "text",
			Description: "description",
			Start:       time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			NotificationParams: NotificationParams{
				Period: time.Hour,
				Params: Params{
					Telegram: 3,
				},
			},
		}
		actual, _ := basicTask.newEvent()

		expected := Event{
			ID:                 0,
			UserID:             2,
			Text:               "text",
			Description:        "description",
			TaskType:           BasicTaskType,
			TaskID:             1,
			NotificationParams: basicTask.NotificationParams,
			LastSent:           time.Time{},
			NextSend:           time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			FirstSend:          time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			Done:               false,
		}

		require.Equal(t, expected, actual)
	})
}
