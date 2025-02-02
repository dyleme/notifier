package jobontime

import (
	"context"
	"sync"
	"time"

	"github.com/Dyleme/Notifier/pkg/utils"
	"github.com/benbjohnson/clock"
)

type JobInTime struct {
	clock        clock.Clock
	timer        *clock.Timer
	sendTimeMx   *sync.RWMutex
	nextSendTime time.Time
	checkPeriod  time.Duration
	job          Job
}

//go:generate	mockgen -destination=job_mocks_test.go -package=jobontime_test . Job
type Job interface {
	GetNextTime(ctx context.Context) (time.Time, bool)
	Do(ctx context.Context, now time.Time)
}

func New(
	clock clock.Clock,
	job Job,
	checkPeriod time.Duration,
) *JobInTime {
	return &JobInTime{
		clock:       clock,
		timer:       clock.Timer(time.Hour),
		sendTimeMx:  &sync.RWMutex{},
		checkPeriod: checkPeriod,
		job:         job,
	}
}

func (j *JobInTime) Run(ctx context.Context) {
	j.setNextEventTime(ctx)
	for {
		select {
		case <-j.timer.C:
			j.job.Do(ctx, j.clock.Now())
			j.setNextEventTime(ctx)
		case <-ctx.Done():
			j.timer.Stop()

			return
		}
	}
}

func (j *JobInTime) setNextEventTime(ctx context.Context) {
	j.sendTimeMx.Lock()
	defer j.sendTimeMx.Unlock()

	t := j.nearestCheckTime(ctx)
	j.nextSendTime = t
	j.timer.Reset(j.clock.Until(t))
}

func (j *JobInTime) nearestCheckTime(ctx context.Context) time.Time {
	nextPeriodicInvocationTime := j.clock.Now().Truncate(time.Minute).Add(j.checkPeriod)
	t, exist := j.job.GetNextTime(ctx)
	if !exist {
		return nextPeriodicInvocationTime
	}

	return utils.MinTime(nextPeriodicInvocationTime, t)
}

func (j *JobInTime) UpdateWithTime(ctx context.Context, t time.Time) {
	j.sendTimeMx.RLock()
	newTimeIsBeforeCurrent := t.Before(j.nextSendTime)
	j.sendTimeMx.RUnlock()

	if newTimeIsBeforeCurrent {
		j.setNextEventTime(ctx)
	}
}
