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
			Tags: []Tag{
				{
					ID:     5,
					UserID: 2,
					Name:   "tag",
				},
			},
		}
		actual, _ := basicTask.newEvent(time.Time{})

		expected := Event{
			ID:                 0,
			UserID:             2,
			Text:               "text",
			Description:        "description",
			TaskType:           BasicTaskType,
			TaskID:             1,
			NextSend:           time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			FirstSend:          time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			Done:               false,
			Notify:             false,
			NotificationParams: basicTask.NotificationParams,
			Tags: []Tag{
				{
					ID:     5,
					UserID: 2,
					Name:   "tag",
				},
			},
		}

		require.Equal(t, expected, actual)
	})
}
