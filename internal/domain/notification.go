package domain

import "time"

type Event struct {
	TaskID             int
	SendingID          int
	Done               bool
	OriginalSending    time.Time
	NextSending        time.Time
	Text               string
	Descriptions       string
	TgID               int
	NotificationPeriod time.Duration
}

func (e Event) ExtractSending() Sending {
	return Sending{
		ID:              e.SendingID,
		TaskID:          e.TaskID,
		Done:            e.Done,
		OriginalSending: e.OriginalSending,
		NextSending:     e.NextSending,
	}
}
