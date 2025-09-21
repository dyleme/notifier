package domain

import "time"

type Notification struct {
	EventID            int
	SendTime           time.Time
	Message            string
	TgID               int
	NotificationPeriod time.Duration
}
