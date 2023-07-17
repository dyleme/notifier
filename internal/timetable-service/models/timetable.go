package models

import (
	"fmt"
	"time"
)

type TimetableTask struct {
	ID          int
	UserID      int
	TaskID      int
	Text        string
	Description string
	Start       time.Time
	Finish      time.Time
	Done        bool
}

func (t TimetableTask) MarkComplete() (TimetableTask, error) {
	if t.Done {
		return TimetableTask{}, fmt.Errorf("already completed")
	}
	t.Done = true

	return t, nil
}

func (t TimetableTask) MarkIncomplete() (TimetableTask, error) {
	if !t.Done {
		return TimetableTask{}, fmt.Errorf("already incompleted")
	}
	t.Done = false

	return t, nil
}
