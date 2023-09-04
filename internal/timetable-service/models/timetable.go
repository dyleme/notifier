package models

import (
	"time"
)

type TimetableTask struct {
	ID           int
	UserID       int
	TaskID       int
	Text         string
	Description  string
	Start        time.Time
	Finish       time.Time
	Done         bool
	Notification Notification
}

type Notification struct {
	Sended bool                `json:"sended"`
	Params *NotificationParams `json:"notification_params"`
}

func (tt *TimetableTask) IsGettingDone(newDone bool) bool {
	if !tt.Done && newDone {
		return true
	}
	return false
}
