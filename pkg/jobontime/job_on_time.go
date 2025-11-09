package jobontime

import (
	"context"
	"sync"
	"time"

	"github.com/benbjohnson/clock"

	"github.com/dyleme/Notifier/pkg/log"
)

type JobInTime struct {
	clock        clock.Clock
	timer        *clock.Timer
	sendTimeMx   *sync.RWMutex
	nextSendTime time.Time
	checkPeriod  time.Duration
	job          Job
}

//go:generate mockgen -destination=job_mocks_test.go -package=jobontime_test . Job
type Job interface {
	GetNextTime(ctx context.Context) (time.Time, bool)
	Do(ctx context.Context, now time.Time)
}

func New(
	cl clock.Clock,
	job Job,
	checkPeriod time.Duration,
) *JobInTime {
	return &JobInTime{
		clock:       cl,
		timer:       cl.Timer(time.Hour),
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
			jobCtx := log.WithCtx(ctx, log.RequestID())
			j.job.Do(jobCtx, j.clock.Now())
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
	log.Ctx(ctx).Debug("set next send time", "time", t)
	j.timer.Reset(j.clock.Until(t))
}

func (j *JobInTime) nearestCheckTime(ctx context.Context) time.Time {
	nextPeriodicInvocationTime := j.clock.Now().Truncate(time.Minute).Add(j.checkPeriod)

	nextEventTime, exist := j.job.GetNextTime(ctx)
	if !exist {
		return nextPeriodicInvocationTime
	}

	return minTime(nextPeriodicInvocationTime, nextEventTime)
}

func minTime(ts ...time.Time) time.Time {
	if len(ts) == 0 {
		return time.Time{}
	}
	minTime := ts[0]
	for _, t := range ts {
		if t.Before(minTime) {
			minTime = t
		}
	}

	return minTime
}

func (j *JobInTime) UpdateWithTime(ctx context.Context, t time.Time) {
	j.sendTimeMx.RLock()
	newTimeIsBeforeCurrent := t.Before(j.nextSendTime)
	j.sendTimeMx.RUnlock()

	if newTimeIsBeforeCurrent {
		j.setNextEventTime(ctx)
	}
}
