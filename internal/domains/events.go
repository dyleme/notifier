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

type SendingEvent struct {
	EventID     int
	UserID      int
	Message     string
	Description string
	Params      NotificationParams
	SendTime    time.Time
}

func NewSendingEvent(n Event, eventParams NotificationParams) SendingEvent {
	return SendingEvent{
		EventID:     n.ID,
		UserID:      n.UserID,
		Message:     n.Text,
		Description: n.Description,
		Params:      eventParams,
		SendTime:    n.SendTime,
	}
}

type Event struct {
	ID          int
	UserID      int
	Text        string
	Description string
	TaskType    TaskType
	TaskID      int
	Params      *NotificationParams
	SendTime    time.Time
	Sended      bool
	Done        bool
}

func (n Event) BelongsTo(userID int) bool {
	return n.UserID == userID
}
