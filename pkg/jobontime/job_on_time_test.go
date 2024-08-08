package jobontime_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/Dyleme/Notifier/pkg/jobontime"
	"github.com/Dyleme/Notifier/pkg/log"
	"github.com/Dyleme/Notifier/pkg/log/mocklogger"
	"github.com/Dyleme/Notifier/pkg/serverrors"
)

const testTimeout = 500 * time.Millisecond

type mock struct {
	job    *MockJob
	clock  *clock.Mock
	logger *mocklogger.MockHandler
}

func setup(t *testing.T) (m mock, check func()) { //nolint:nonamedreturns // for readability
	t.Helper()
	mckCtrl := gomock.NewController(t)
	nower := clock.NewMock()
	nower.Set(time.Now().Truncate(time.Minute))
	m = mock{
		job:    NewMockJob(mckCtrl),
		clock:  nower,
		logger: mocklogger.NewHandler(),
	}

	return m, mckCtrl.Finish
}

var errEventNotFound = serverrors.NewNotFoundError(errors.New("not found"), "event")

func TestNotificationJob_Run(t *testing.T) {
	t.Parallel()
	t.Run("check first event on start up", func(t *testing.T) {
		t.Parallel()
		mock, finish := setup(t)
		defer finish()
		nextInvocation := mock.clock.Now().Add(100 * time.Millisecond)
		mock.job.EXPECT().GetNextTime(gomock.Any()).Return(nextInvocation)
		mock.job.EXPECT().Do(gomock.Any(), nextInvocation)
		mock.job.EXPECT().GetNextTime(gomock.Any()).Return(time.Time{})
		ctx := context.Background()
		ctx = log.InCtx(ctx, slog.New(mock.logger))
		job := jobontime.New(mock.clock, mock.job, 5*time.Minute)
		ctx, cancel := context.WithTimeout(ctx, testTimeout)
		wait := make(chan struct{})
		go func() {
			job.Run(ctx)
			wait <- struct{}{}
		}()
		time.Sleep(time.Millisecond)
		mock.clock.Add(time.Minute)
		cancel()
		<-wait
		require.NoError(t, mock.logger.Error(), "error logged")
	})
	t.Run("invocate if there is no events", func(t *testing.T) {
		t.Parallel()
		mock, finish := setup(t)
		defer finish()
		checkPeriod := 5 * time.Minute
		mock.job.EXPECT().GetNextTime(gomock.Any()).Return(time.Time{})
		mock.job.EXPECT().Do(gomock.Any(), mock.clock.Now().Add(checkPeriod))
		mock.job.EXPECT().GetNextTime(gomock.Any()).Return(time.Time{})

		ctx := log.InCtx(context.Background(), slog.New(mock.logger))
		job := jobontime.New(mock.clock, mock.job, checkPeriod)
		ctx, cancel := context.WithTimeout(ctx, testTimeout)
		wait := make(chan struct{})
		go func() {
			job.Run(ctx)
			wait <- struct{}{}
		}()
		time.Sleep(time.Millisecond)
		mock.clock.Add(checkPeriod + time.Second)
		cancel()
		<-wait
		require.NoError(t, mock.logger.Error(), "error logged")
	})
}
