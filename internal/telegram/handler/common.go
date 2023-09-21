package handler

import (
	"time"
)

const (
	timeDoublePointsFormat = "15:04"
	timeSpaceFormat        = "15 04"
)
const day = 24 * time.Hour

var timeFormats = []string{timeDoublePointsFormat, timeSpaceFormat}

func parseTime(dayString string) (time.Time, error) {
	for _, format := range timeFormats {
		t, err := time.Parse(format, dayString)
		if err == nil { // err eq nil
			return t, nil
		}
	}

	return time.Time{}, ErrCantParseMessage
}

const (
	dayPointFormat         = "02.01"
	daySpaceFormat         = "02 01"
	dayPointWithYearFormat = "02.01.2006"
	daySpaceWithYearFormat = "02 01 2006"
)

var dayFormats = []string{dayPointFormat, daySpaceFormat, dayPointWithYearFormat, daySpaceWithYearFormat}

func parseDay(dayString string) (time.Time, error) {
	for _, format := range dayFormats {
		t, err := time.Parse(format, dayString)
		if err != nil {
			continue
		}

		if t.Year() == 0 {
			t = t.AddDate(time.Now().Year(), 0, 0)
			if t.Before(time.Now()) {
				t = t.AddDate(1, 0, 0)
			}
		}

		return t, nil
	}

	return time.Time{}, ErrCantParseMessage
}
