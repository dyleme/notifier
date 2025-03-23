package domain

import (
	"time"

	"github.com/Dyleme/Notifier/internal/domain/apperr"
	utils "github.com/Dyleme/Notifier/pkg/utils/ptr"
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
		SendTime:    ev.NextSend,
	}
}

type Event struct {
	ID                 int
	UserID             int
	Text               string
	Description        string
	TaskType           TaskType
	TaskID             int
	NextSend           time.Time
	FirstSend          time.Time
	Done               bool
	Notify             bool
	NotificationParams NotificationParams
	Tags               []Tag
}

func (ev Event) BelongsTo(userID int) error {
	if ev.UserID == userID {
		return nil
	}

	return apperr.NewNotBelongToUserError("event", ev.ID, ev.UserID, userID)
}

func (ev Event) Rescheule(now time.Time) Event {
	return ev.RescheuleToTime(now.Add(ev.NotificationParams.Period))
}

func (ev Event) RescheuleToTime(t time.Time) Event {
	ev.NextSend = t

	return ev
}

func (ev Event) MarkDone() Event {
	ev.Done = true

	return ev
}

func (ev Event) NewNotification() (Notification, error) {
	if err := ev.Validate(); err != nil {
		return Notification{}, err
	}

	return Notification{
		EventID:  ev.ID,
		SendTime: ev.NextSend,
		Message:  ev.Text,
		Params:   ev.NotificationParams,
	}, nil
}

func (ev Event) Validate() error {
	if ev.Notify && utils.IsZero(ev.NotificationParams) {
		return apperr.UnexpectedStateError{
			Object: "event",
			Reason: "mark as being notified but notification params are empty",
		}
	}

	if !ev.Notify && !utils.IsZero(ev.NotificationParams) {
		return apperr.UnexpectedStateError{
			Object: "event",
			Reason: "mark as not being notified but notification params exists",
		}
	}

	return nil
}
