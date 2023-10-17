package utils_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Dyleme/Notifier/pkg/utils"
)

func TestSmallestTimeWithoutZero(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		a      time.Time
		b      time.Time
		result int
	}{
		{
			name:   "only zero times",
			a:      time.Time{},
			b:      time.Time{},
			result: 0,
		},
		{
			name:   "first time is zero",
			a:      time.Time{},
			b:      time.Date(2023, 10, 17, 23, 27, 0, 0, time.UTC),
			result: 1,
		},
		{
			name:   "second time is zero",
			a:      time.Date(2023, 10, 17, 23, 27, 0, 0, time.UTC),
			b:      time.Time{},
			result: -1,
		},
		{
			name:   "first time smaller",
			a:      time.Date(2023, 10, 17, 23, 27, 0, 0, time.UTC),
			b:      time.Date(2024, 10, 17, 23, 27, 0, 0, time.UTC),
			result: -1,
		},
		{
			name:   "second time smaller",
			a:      time.Date(2024, 10, 17, 23, 27, 0, 0, time.UTC),
			b:      time.Date(2023, 10, 17, 23, 27, 0, 0, time.UTC),
			result: 1,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.result, utils.TimeCmpWithoutZero(tt.a, tt.b))
		})
	}
}
