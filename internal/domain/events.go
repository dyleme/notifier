package domain

import (
	"encoding/json"
	"time"
)

type NotificationParams struct {
	Period time.Duration `json:"period"`
	Params Params        `json:"params"`
}

func (np *NotificationParams) JSON() []byte {
	bts, err := json.Marshal(np)
	if err != nil {
		panic(err)
	}

	return bts
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

// func NewSendingEvent(ev Event) SendingEvent {
// 	return SendingEvent{
// 		EventID:     ev.ID,
// 		UserID:      ev.UserID,
// 		Message:     ev.Text,
// 		Description: ev.Description,
// 		Params:      ev.NotificationParams,
// 		SendTime:    ev.NextSend,
// 	}
// }
