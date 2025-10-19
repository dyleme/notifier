package domain

import (
	"fmt"
	"time"

	"github.com/dyleme/Notifier/internal/domain/apperr"
)

type SingleTask struct {
	taskCore
	Date time.Time
}

const (
	dateKey TaskParamKey = "date"
)

func (st SingleTask) BuildTask() Task {
	params := map[TaskParamKey]any{
		dateKey: st.Date,
	}

	task := st.Task(params)

	return task
}

func ParseSingleTask(t Task) (SingleTask, error) {
	dateAny, ok := t.Params[dateKey]
	if !ok {
		return SingleTask{}, fmt.Errorf("missing date : %w", apperr.ErrInternal)
	}

	dateStr, ok := dateAny.(string)
	if !ok {
		return SingleTask{}, fmt.Errorf("date is not string [%v]: %w", dateAny, apperr.ErrInternal)
	}

	date, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return SingleTask{}, fmt.Errorf("parse date [%v]: %w", dateStr, err)
	}

	st := SingleTask{
		taskCore: t.core(),
		Date:     date,
	}

	return st, nil
}

func NewSingleTask(t TaskCreationParams, date time.Time) SingleTask {
	return SingleTask{
		taskCore: taskCore{
			ID:          t.ID,
			Text:        t.Text,
			Description: t.Description,
			UserID:      t.UserID,
			Type:        Single,
			Start:       t.Start,
		},
		Date: date,
	}
}

func (st SingleTask) NewEvent() Sending {
	return Sending{
		TaskID:          st.ID,
		Done:            false,
		OriginalSending: st.Date.Add(st.Start),
		NextSending:     st.Date.Add(st.Start),
	}
}
