package domains

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPeriodicTask_newEvent(t *testing.T) {
	t.Parallel()
	t.Run("check mapping", func(t *testing.T) {
		t.Parallel()
		periodicTask := PeriodicTask{
			ID:             1,
			Text:           "text",
			Description:    "description",
			UserID:         2,
			Start:          3 * time.Hour,
			SmallestPeriod: 24 * time.Hour,
			BiggestPeriod:  3 * 24 * time.Hour,
			NotificationParams: NotificationParams{
				Period: time.Hour,
				Params: Params{Telegram: 3},
			},
		}
		actual, err := periodicTask.newEvent()
		require.NoError(t, err)

		expected := Event{
			ID:                 0,
			UserID:             2,
			Text:               "text",
			Description:        "description",
			TaskType:           PeriodicTaskType,
			TaskID:             1,
			NotificationParams: periodicTask.NotificationParams,
			LastSent:           time.Time{},
			NextSend:           time.Time{},
			FirstSend:          time.Time{},
			Done:               false,
		}

		// do not check send time
		actual.NextSend = time.Time{}
		actual.FirstSend = time.Time{}
		expected.NextSend = time.Time{}
		expected.FirstSend = time.Time{}
		require.Equal(t, expected, actual)
	})

	timeDay := 24 * time.Hour
	nowTime := time.Now()
	dayBeginning := nowTime.Truncate(timeDay)
	testCases := []struct {
		name       string
		pt         PeriodicTask
		now        time.Time
		isError    bool
		highBorder time.Time
		lowBorder  time.Time
	}{
		{
			name: "high border lower than low border",
			pt: PeriodicTask{
				Start:          time.Hour,
				SmallestPeriod: 3 * timeDay,
				BiggestPeriod:  timeDay,
			},
			now:     nowTime,
			isError: true,
		},
		{
			name: "check time in period",
			pt: PeriodicTask{
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
			pt: PeriodicTask{
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
				actualEvent, err := tc.pt.newEvent()
				actual := actualEvent.NextSend

				require.Equalf(t, actualEvent.NextSend, actualEvent.FirstSend,
					"next send time[%v] not equal first send time[%v]", actualEvent.NextSend, actualEvent.FirstSend)
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
