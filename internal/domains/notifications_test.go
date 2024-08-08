package domains_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Dyleme/Notifier/internal/domains"
)

func TestEvent_NewSendingEvent(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name      string
		event     domains.Event
		expected  domains.Notification
		expectErr bool
	}{
		{
			name: "mapping notified event",
			event: domains.Event{
				ID:                 1,
				UserID:             2,
				Text:               "text",
				Description:        "description",
				TaskType:           domains.BasicTaskType,
				TaskID:             3,
				LastSendedTime:     time.Time{},
				Time:               time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				FirstTime:          time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				Notify:             true,
				Done:               false,
				NotificationParams: &domains.NotificationParams{Period: time.Minute, Params: domains.Params{Telegram: 5}},
				Tags:               []domains.Tag{},
			},
			expected: domains.Notification{
				EventID:     1,
				UserID:      2,
				Message:     "text",
				Description: "description",
				Params:      domains.NotificationParams{Period: time.Minute, Params: domains.Params{Telegram: 5}},
				SendTime:    time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "mapping unnotified event",
			event: domains.Event{
				ID:                 1,
				UserID:             2,
				Text:               "text",
				Description:        "description",
				TaskType:           domains.BasicTaskType,
				TaskID:             3,
				LastSendedTime:     time.Time{},
				Time:               time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				FirstTime:          time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				Notify:             false,
				Done:               false,
				NotificationParams: nil,
				Tags:               []domains.Tag{},
			},
			expectErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			actual, err := tc.event.NewNotification()

			require.Equalf(t, tc.expectErr, err != nil, "expect error: %v, actual: %v", tc.expectErr, err != nil)
			require.Equal(t, tc.expected, actual)
		})
	}
}
