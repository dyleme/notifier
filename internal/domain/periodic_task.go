package domain

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/Dyleme/Notifier/internal/domain/apperr"
)

const timeDay = 24 * time.Hour

type InvalidPeriodTimeError struct {
	smallest time.Duration
	biggest  time.Duration
}

var ErrNotificaitonParamsRequired = errors.New("notification params is required")

func (i InvalidPeriodTimeError) Error() string {
	return fmt.Sprintf("invalid period error biggest is before smallest %v < %v", i.biggest, i.smallest)
}

const (
	smallestPeriodKey TaskParamKey = "smallest_period"
	biggestPeriodKey  TaskParamKey = "biggest_period"
)

type PeriodicTask struct {
	taskCore
	SmallestPeriod time.Duration
	BiggestPeriod  time.Duration
}

func ParsePeriodicTask(t Task) (PeriodicTask, error) {
	smallestPeriodAny, ok := t.Params[smallestPeriodKey]
	if !ok {
		return PeriodicTask{}, fmt.Errorf("missing smallest period : %w", apperr.ErrInternal)
	}

	smallestPeriodFloat, ok := smallestPeriodAny.(float64)
	if !ok {
		return PeriodicTask{}, fmt.Errorf("smallest period is not float [%v]: %w", smallestPeriodAny, apperr.ErrInternal)
	}

	biggestPeriodAny, ok := t.Params[biggestPeriodKey]
	if !ok {
		return PeriodicTask{}, fmt.Errorf("missing biggest period : %w", apperr.ErrInternal)
	}

	biggestPeriodFloat, ok := biggestPeriodAny.(float64)
	if !ok {
		return PeriodicTask{}, fmt.Errorf("biggest period is not float [%v]: %w", biggestPeriodAny, apperr.ErrInternal)
	}

	periodicTask := PeriodicTask{
		taskCore:       t.core(),
		SmallestPeriod: time.Duration(smallestPeriodFloat),
		BiggestPeriod:  time.Duration(biggestPeriodFloat),
	}

	return periodicTask, nil
}

func (pt PeriodicTask) BuildTask() Task {
	params := map[TaskParamKey]any{
		smallestPeriodKey: pt.SmallestPeriod,
		biggestPeriodKey:  pt.BiggestPeriod,
	}

	t := pt.Task(params)

	return t
}

func NewPeriodicTask(params TaskCreationParams, smallestPeriod, biggestPeriod time.Duration) PeriodicTask {
	return PeriodicTask{
		taskCore: taskCore{
			ID:          params.ID,
			Text:        params.Text,
			Description: params.Description,
			UserID:      params.UserID,
			Type:        Periodic,
			Start:       params.Start,
		},
		SmallestPeriod: smallestPeriod,
		BiggestPeriod:  biggestPeriod,
	}
}

func (pt PeriodicTask) NewSending(now time.Time) Sending {
	minDays := int(pt.SmallestPeriod / timeDay)
	maxDays := int(pt.BiggestPeriod / timeDay)
	days := minDays
	if diff := maxDays - minDays; diff > 0 {
		days += rand.IntN(diff) //nolint:gosec // no need for security
	}
	dayBeginning := now.Add(time.Duration(days) * timeDay).Truncate(timeDay)
	sendTime := dayBeginning.Add(pt.Start)

	return Sending{
		TaskID:          pt.ID,
		Done:            false,
		OriginalSending: sendTime,
		NextSending:     sendTime,
	}
}

func (pt PeriodicTask) TimeParamsHasChanged(updT PeriodicTask) bool {
	return pt.SmallestPeriod != updT.SmallestPeriod ||
		pt.BiggestPeriod != updT.BiggestPeriod ||
		pt.Start != updT.Start
}
