package notification

import "time"

type Notification struct {
	ID               int
	UserID           int
	Message          string
	NotificationTime time.Time
	Destinations     []string
	TaskID           int
}
