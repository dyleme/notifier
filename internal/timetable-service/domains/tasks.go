package domains

import (
	"time"
)

type Task struct {
	ID           int
	UserID       int
	Text         string
	RequiredTime time.Duration
	Periodic     bool
	Done         bool
	Archived     bool
}

func (t *Task) UsedTask() Task {
	if !t.Periodic {
		t.Archived = true
	}
	return *t
}

func (t *Task) CanUse() bool {
	return !t.Archived
}
