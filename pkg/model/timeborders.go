package model

import "time"

type TimeBorders struct {
	From time.Time
	To   time.Time
}

var (
	minTime = time.Unix(0, 0)
	maxTime = time.Unix(1<<63-1, 0)
)

func New(lowerBorder, upperBorder time.Time) TimeBorders {
	return TimeBorders{From: lowerBorder, To: upperBorder}
}

func NewInfiniteUpper(lowerBorder time.Time) TimeBorders {
	return TimeBorders{From: lowerBorder, To: maxTime}
}

func NewInfiniteLower(upperBorder time.Time) TimeBorders {
	return TimeBorders{From: minTime, To: upperBorder}
}

func NewInfinite() TimeBorders {
	return TimeBorders{From: minTime, To: maxTime}
}
