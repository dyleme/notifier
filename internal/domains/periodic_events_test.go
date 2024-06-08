package domains_test

import (
	"testing"
	"time"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/stretchr/testify/require"
)

func TestPeriodicEvent_NewNotification(t *testing.T) {
	t.Parallel()
	t.Run("check mapping", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		periodicEvent := domains.PeriodicEvent{
			ID:             1,
			Text:           "text",
			Description:    "description",
			UserID:         2,
			Start:          3 * time.Hour,
			SmallestPeriod: 24 * time.Hour,
			BiggestPeriod:  3 * 24 * time.Hour,
			NotificationParams: &domains.NotificationParams{
				Period: time.Hour,
				Params: domains.Params{Telegram: 3},
			},
		}
		actual, err := periodicEvent.NewNotification(now)
		require.NoError(t, err)

		expected := domains.Notification{
			ID:          0,
			UserID:      2,
			Text:        "text",
			Description: "description",
			EventType:   domains.PeriodicEventType,
			EventID:     1,
			Params:      periodicEvent.NotificationParams,
			Sended:      false,
			Done:        false,
		}

		// do not check send time
		actual.SendTime = time.Time{}
		expected.SendTime = time.Time{}
		require.Equal(t, expected, actual)
	})

	timeDay := 24 * time.Hour
	nowTime := time.Now()
	dayBeginning := nowTime.Truncate(timeDay)
	testCases := []struct {
		name       string
		pe         domains.PeriodicEvent
		now        time.Time
		isError    bool
		highBorder time.Time
		lowBorder  time.Time
	}{
		{
			name: "high border lower than low border",
			pe: domains.PeriodicEvent{
				Start:          time.Hour,
				SmallestPeriod: 3 * timeDay,
				BiggestPeriod:  timeDay,
			},
			now:     nowTime,
			isError: true,
		},
		{
			name: "check time in period",
			pe: domains.PeriodicEvent{
				Start:          2 * time.Hour,
				SmallestPeriod: timeDay,
				BiggestPeriod:  11 * timeDay,
			},
			now:        nowTime,
			lowBorder:  dayBeginning.Add(timeDay),
			highBorder: dayBeginning.Add(12 * timeDay),
		},
		{
			name: "smallest period equal to biggest period",
			pe: domains.PeriodicEvent{
				Start:          3 * time.Hour,
				SmallestPeriod: timeDay,
				BiggestPeriod:  timeDay,
			},
			now:        nowTime,
			lowBorder:  dayBeginning.Add(timeDay),
			highBorder: dayBeginning.Add(2 * timeDay),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			for range 10 {
				notif, err := tc.pe.NewNotification(tc.now)
				actual := notif.SendTime

				require.True(t, tc.isError == (err != nil), "check error")
				if tc.isError != (err != nil) {
					t.Errorf("[waiting err = %v, actualError=%v]", tc.isError, err)
				}
				if actual.Before(tc.lowBorder) || actual.After(tc.highBorder) {
					t.Errorf("send time should be in range [generatedTime=%v, lowBorder=%v, highBorder=%v]", actual, tc.lowBorder, tc.highBorder)
				}
			}
		})
	}
}
