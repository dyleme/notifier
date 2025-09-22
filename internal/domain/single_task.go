package domain

import (
	"time"
)

type SingleTask struct {
	Task
}

const (
	dateKey CreationParamKey = "date"
)

func (st SingleTask) Date() time.Time {
	dateStr := st.EventCreationParams[dateKey].(string)

	t, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		panic(err)
	}

	return t
}

func NewSingleTask(t TaskCreationParams, date time.Time) SingleTask {
	return SingleTask{
		Task: Task{
			ID:          t.ID,
			Text:        t.Text,
			Description: t.Description,
			UserID:      t.UserID,
			Type:        Single,
			Start:       t.Start,
			EventCreationParams: map[CreationParamKey]any{
				dateKey: date.Format(time.RFC3339),
			},
		},
	}
}

func (st SingleTask) NewEvent() Sending {
	return Sending{
		TaskID:          st.ID,
		Done:            false,
		OriginalSending: st.Date().Add(st.Start),
		NextSending:     st.Date().Add(st.Start),
	}
}
