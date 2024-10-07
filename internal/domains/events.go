package domains

import (
	"time"
)

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
		SendTime:    ev.NextSendTime,
	}
}

type Event struct {
	ID                 int
	UserID             int
	Text               string
	Description        string
	TaskType           TaskType
	TaskID             int
	LastSendedTime     time.Time
	NextSendTime       time.Time
	FirstSendTime      time.Time
	Done               bool
	NotificationParams NotificationParams
}

func (ev Event) BelongsTo(userID int) error {
	if ev.UserID == userID {
		return nil
	}

	return NewNotBelongToUserError("event", ev.ID, ev.UserID, userID)
}

func (ev Event) Rescheule(now time.Time) Event {
	return ev.RescheuleToTime(now.Add(ev.NotificationParams.Period))
}

func (ev Event) RescheuleToTime(t time.Time) Event {
	ev.NextSendTime = t

	return ev
}

func (ev Event) MarkDone() Event {
	ev.Done = true

	return ev
}
