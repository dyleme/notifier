package domain

import (
	"time"
)

type TaskType string

const (
	Periodic TaskType = "periodic"
	Single   TaskType = "single"
)

type CreationParamKey string

type Task struct {
	ID                  int
	CreatedAt           time.Time
	Text                string
	Description         string
	UserID              int
	Type                TaskType
	Start               time.Duration
	EventCreationParams map[CreationParamKey]any
}

type TaskCreationParams struct {
	ID          int
	Text        string
	Description string
	UserID      int
	Start       time.Duration
}

type Sending struct {
	ID              int
	CreatedAt       time.Time
	TaskID          int
	Done            bool
	OriginalSending time.Time
	NextSending     time.Time
}

func (s Sending) Rescheule(now time.Time, period time.Duration) Sending {
	return s.RescheuleToTime(now.Add(period))
}

func (s Sending) RescheuleToTime(t time.Time) Sending {
	s.NextSending = t

	return s
}

func (s Sending) MarkDone() Sending {
	s.Done = true

	return s
}
