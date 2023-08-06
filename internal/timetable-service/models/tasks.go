package models

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

func (t *Task) ToTimetableTask(start time.Time, description string) TimetableTask {
	return TimetableTask{ //nolint:exhaustruct  // TODO: We dont know timetabletask id
		UserID:      t.UserID,
		Text:        t.Text,
		Start:       start,
		Finish:      start.Add(t.RequiredTime),
		Done:        false,
		TaskID:      t.ID,
		Description: description,
	}
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
