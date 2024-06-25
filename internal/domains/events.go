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

func NewSendingEvent(ev Event) SendingEvent {
	return SendingEvent{
		EventID:     ev.ID,
		UserID:      ev.UserID,
		Message:     ev.Text,
		Description: ev.Description,
		Params:      ev.NotificationParams,
		SendTime:    ev.SendTime,
	}
}

type Event struct {
	ID                 int
	UserID             int
	Text               string
	Description        string
	TaskType           TaskType
	TaskID             int
	NotificationParams NotificationParams
	SendTime           time.Time
	Sended             bool
	Done               bool
}

func (ev Event) BelongsTo(userID int) error {
	if ev.UserID == userID {
		return nil
	}

	return NewNotBelongToUserError("event", ev.ID, ev.UserID, userID)
}
