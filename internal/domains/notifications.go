package domains

import "time"

type NotificationParams struct {
	Period time.Duration `json:"period"`
	Params Params        `json:"params"`
}

type Params struct {
	Telegram int    `json:"telegram,omitempty"`
	Webhook  string `json:"webhook,omitempty"`
	Cmd      string `json:"cmd,omitempty"`
}

type SendingNotification struct {
	EventType        EventType
	EventID          int
	UserID           int
	Message          string
	Description      string
	Params           NotificationParams
	NotificationTime time.Time
}
