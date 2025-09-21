package domain

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"time"
)

type TaskType string

const (
	Periodic TaskType = "periodic"
	Signle   TaskType = "single"
)

type Task struct {
	ID                  int
	CreatedAt           time.Time
	Text                string
	Description         string
	UserID              int
	Type                TaskType
	Start               time.Duration
	EventCreationParams json.RawMessage
}

type PeriodicTask struct {
	Task
	SmallestPeriod time.Duration `json:"smallest_period"`
	BiggestPeriod  time.Duration `json:"biggest_period"`
}

func PeriodictaskFromTask(t Task) (PeriodicTask, error) {
	var params PeriodicTask
	err := json.Unmarshal(t.EventCreationParams, &params)
	if err != nil {
		return PeriodicTask{}, fmt.Errorf("unmarshal: %w", err)
	}

	return PeriodicTask{
		Task:           t,
		SmallestPeriod: params.SmallestPeriod,
		BiggestPeriod:  params.BiggestPeriod,
	}, nil
}

func (pt PeriodicTask) NewEvent(now time.Time) Event {
	minDays := int(pt.SmallestPeriod / timeDay)
	maxDays := int(pt.BiggestPeriod / timeDay)
	days := minDays
	if diff := maxDays - minDays; diff > 0 {
		days += rand.IntN(diff)
	}
	dayBeginning := now.Add(time.Duration(days) * timeDay).Truncate(timeDay)
	sendTime := dayBeginning.Add(pt.Start)

	return Event{
		TaskID:          pt.ID,
		Done:            false,
		OriginalSending: sendTime,
		NextSending:     sendTime,
	}
}

type SingleTask struct {
	Task
	Date time.Time `json:"date"`
}

func SingleTaskFromTask(t Task) (SingleTask, error) {
	var params SingleTask
	err := json.Unmarshal(t.EventCreationParams, &params)
	if err != nil {
		return SingleTask{}, fmt.Errorf("unmarshal: %w", err)
	}
	return SingleTask{
		Task: t,
		Date: params.Date,
	}, nil
}

func (st SingleTask) NewEvent() Event {
	return Event{
		TaskID:          st.ID,
		Done:            false,
		OriginalSending: st.Date.Add(st.Start),
		NextSending:     st.Date.Add(st.Start),
	}
}

type Event struct {
	ID              int
	CreatedAt       time.Time
	TaskID          int
	Done            bool
	OriginalSending time.Time
	NextSending     time.Time
}

func (ev Event) Rescheule(now time.Time, period time.Duration) Event {
	return ev.RescheuleToTime(now.Add(period))
}

func (ev Event) RescheuleToTime(t time.Time) Event {
	ev.NextSending = t

	return ev
}

func (ev Event) MarkDone() Event {
	ev.Done = true

	return ev
}
