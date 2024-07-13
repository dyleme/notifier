package utils

import "time"

func MinTime(ts ...time.Time) time.Time {
	if len(ts) == 0 {
		return time.Time{}
	}
	minTime := ts[0]
	for _, t := range ts {
		if t.Before(minTime) {
			minTime = t
		}
	}

	return minTime
}
