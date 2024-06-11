package domains_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Dyleme/Notifier/internal/domains"
)

func TestBasicTask_NewEvent(t *testing.T) {
	t.Parallel()
	t.Run("check mapping", func(t *testing.T) {
		t.Parallel()
		basicTask := domains.BasicTask{
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
		actual := basicTask.NewEvent()

		expected := domains.Event{
			ID:          0,
			UserID:      2,
			Text:        "text",
			Description: "description",
			TaskType:    domains.BasicTaskType,
			TaskID:      1,
			Params:      basicTask.NotificationParams,
			SendTime:    time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			Sended:      false,
			Done:        false,
		}

		require.Equal(t, expected, actual)
	})
}
