package domains

import (
	"time"

	"github.com/friendsofgo/errors"
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

type Notification struct {
	EventID     int
	UserID      int
	Message     string
	Description string
	Params      NotificationParams
	SendTime    time.Time
}

type DailyNotification struct {
	ToDo    []DailyNotificationEvent
	NotDone []DailyNotificationEvent
}
type DailyNotificationEvent struct {
	EventID     int
	UserID      int
	Message     string
	Description string
	Time        time.Time
}

var ErrNoNotifiedEvent = errors.New("not notified event")

func (ev Event) NewNotification() (Notification, error) {
	if !ev.Notify {
		return Notification{}, ErrNoNotifiedEvent
	}
	err := ev.Validate()
	if err != nil {
		return Notification{}, err
	}
	return Notification{
		EventID:     ev.ID,
		UserID:      ev.UserID,
		Message:     ev.Text,
		Description: ev.Description,
		Params:      *ev.NotificationParams,
		SendTime:    ev.Time,
	}, nil
}

func (ev Event) NewDailyNotificationEvent() DailyNotificationEvent {
	return DailyNotificationEvent{
		EventID:     ev.ID,
		UserID:      ev.UserID,
		Message:     ev.Text,
		Description: ev.Description,
		Time:        ev.Time,
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
	Time               time.Time
	FirstTime          time.Time
	Notify             bool
	Done               bool
	NotificationParams *NotificationParams
	Tags               []Tag
}

func (ev Event) Validate() error {
	if ev.Notify && ev.NotificationParams == nil {
		return ErrNotificaitonParamsRequired
	}

	return nil
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
	ev.Time = t

	return ev
}
