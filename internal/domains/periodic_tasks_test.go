package domains_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/Dyleme/Notifier/internal/domains"
)

func TestPeriodicTask_NewEvent(t *testing.T) {
	t.Parallel()
	t.Run("check mapping", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		periodicTask := domains.PeriodicTask{
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
		actual, err := periodicTask.NewEvent(now)
		require.NoError(t, err)

		expected := domains.Event{
			ID:          0,
			UserID:      2,
			Text:        "text",
			Description: "description",
			TaskType:    domains.PeriodicTaskType,
			TaskID:      1,
			Params:      periodicTask.NotificationParams,
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
		pt         domains.PeriodicTask
		now        time.Time
		isError    bool
		highBorder time.Time
		lowBorder  time.Time
	}{
		{
			name: "high border lower than low border",
			pt: domains.PeriodicTask{
				Start:          time.Hour,
				SmallestPeriod: 3 * timeDay,
				BiggestPeriod:  timeDay,
			},
			now:     nowTime,
			isError: true,
		},
		{
			name: "check time in period",
			pt: domains.PeriodicTask{
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
			pt: domains.PeriodicTask{
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
				event, err := tc.pt.NewEvent(tc.now)
				actual := event.SendTime

				require.Equal(t, tc.isError, err != nil, "check error")
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
