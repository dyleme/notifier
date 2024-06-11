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
	NotificationID int
	UserID         int
	Message        string
	Description    string
	Params         NotificationParams
	SendTime       time.Time
}

func NewSendingNotification(n Notification, notificationParams NotificationParams) SendingNotification {
	return SendingNotification{
		NotificationID: n.ID,
		UserID:         n.UserID,
		Message:        n.Text,
		Description:    n.Description,
		Params:         notificationParams,
		SendTime:       n.SendTime,
	}
}

type Notification struct {
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

func (n Notification) BelongsTo(userID int) bool {
	return n.UserID == userID
}
