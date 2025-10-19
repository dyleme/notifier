package domain

import (
	"testing"
	"time"
)

const day = 24 * time.Hour

func TestPeriodicTask_NewSending(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		pt   PeriodicTask
		now  time.Time
	}{
		{
			name: "happy path",
			pt: PeriodicTask{
				SmallestPeriod: 2 * day,
				BiggestPeriod:  7 * day,
				taskCore: taskCore{
					Text:        "text",
					Description: "description",
					Start:       time.Hour,
				},
			},
			now: time.Date(2023, 10, 11, 22, 11, 0, 0, time.UTC),
		},
		{
			name: "equal period",
			pt: PeriodicTask{
				SmallestPeriod: 7 * day,
				BiggestPeriod:  7 * day,
				taskCore: taskCore{
					Text:        "text",
					Description: "description",
					Start:       time.Hour,
				},
			},
			now: time.Date(2023, 10, 11, 22, 11, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sending := tt.pt.NewSending(tt.now)
			start := tt.now.Truncate(day).Add(tt.pt.Start)
			if sending.OriginalSending.Before(start.Add(tt.pt.SmallestPeriod)) {
				t.Errorf("original sending [%v] is before smallest period [%v] now[%v]", sending.OriginalSending, tt.pt.SmallestPeriod, tt.now)
			}

			if sending.OriginalSending.After(start.Add(tt.pt.BiggestPeriod)) {
				t.Errorf("original sending [%v] is after biggest period [%v] now[%v]", sending.OriginalSending, tt.pt.BiggestPeriod, tt.now)
			}
		})
	}
}
