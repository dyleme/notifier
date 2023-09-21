package handler

import (
	"testing"
	"time"
)

func Test_getTimezone(t *testing.T) {
	tests := []struct {
		name        string
		utcTime     time.Time
		userHours   int
		userMinutes int
		want        int
	}{
		{
			name:        "UTC",
			utcTime:     time.Date(2023, 10, 11, 22, 11, 0, 0, time.UTC),
			userHours:   22,
			userMinutes: 11,
			want:        0,
		},
		{
			name:        "UTC-01",
			utcTime:     time.Date(2023, 10, 11, 22, 11, 0, 0, time.UTC),
			userHours:   21,
			userMinutes: 11,
			want:        -1,
		},
		{
			name:        "UTC+01",
			utcTime:     time.Date(2023, 10, 11, 22, 11, 0, 0, time.UTC),
			userHours:   23,
			userMinutes: 11,
			want:        1,
		},
		{
			name:        "UTC+04 trough noon up",
			utcTime:     time.Date(2023, 10, 11, 22, 11, 0, 0, time.UTC),
			userHours:   2,
			userMinutes: 11,
			want:        4,
		},
		{
			name:        "UTC+04 trough noon down",
			utcTime:     time.Date(2023, 10, 11, 2, 11, 0, 0, time.UTC),
			userHours:   22,
			userMinutes: 11,
			want:        -4,
		},
		{
			name:        "UTC+04 less minutes",
			utcTime:     time.Date(2023, 10, 11, 2, 11, 0, 0, time.UTC),
			userHours:   22,
			userMinutes: 13,
			want:        -4,
		},
		{
			name:        "UTC+04 more minutes",
			utcTime:     time.Date(2023, 10, 11, 2, 11, 0, 0, time.UTC),
			userHours:   22,
			userMinutes: 8,
			want:        -4,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getTimezone(tt.utcTime, tt.userHours, tt.userMinutes); got != tt.want {
				t.Errorf("getTimezone() = %v, want %v", got, tt.want)
			}
		})
	}
}
