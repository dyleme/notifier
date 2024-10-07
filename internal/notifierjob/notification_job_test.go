package notifierjob_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/internal/notifierjob"
	"github.com/Dyleme/Notifier/internal/notifierjob/mocks"
	"github.com/Dyleme/Notifier/pkg/log"
	"github.com/Dyleme/Notifier/pkg/log/mocklogger"
	"github.com/Dyleme/Notifier/pkg/serverrors"
	"github.com/Dyleme/Notifier/pkg/testutils"
)

const testTimeout = 500 * time.Millisecond

type mock struct {
	notifier *mocks.MockNotifier
	repo     *mocks.MockRepository
	clock    *clock.Mock
	logger   *mocklogger.MockHandler
}

func setup(t *testing.T) (m mock, check func()) { //nolint:nonamedreturns // for readability
	t.Helper()
	mckCtrl := gomock.NewController(t)
	nower := clock.NewMock()
	nower.Set(time.Now().Truncate(time.Minute))
	m = mock{
		notifier: mocks.NewMockNotifier(mckCtrl),
		repo:     mocks.NewMockRepository(mckCtrl),
		clock:    nower,
		logger:   mocklogger.NewHandler(),
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
		mock.repo.EXPECT().GetNearest(gomock.Any()).Return(domains.Event{NextSendTime: nextInvocation}, nil)
		mock.repo.EXPECT().ListNotSended(gomock.Any(), nextInvocation).Return([]domains.Event{}, nil)
		mock.repo.EXPECT().GetNearest(gomock.Any()).Return(domains.Event{}, errEventNotFound)
		ctx := context.Background()
		ctx = log.InCtx(ctx, slog.New(mock.logger))
		nj := notifierjob.New(mock.repo, notifierjob.Config{CheckTasksPeriod: time.Minute}, testutils.MockTXManager{}, mock.clock)
		nj.SetNotifier(mock.notifier)
		ctx, cancel := context.WithTimeout(ctx, testTimeout)
		wait := make(chan struct{})
		go func() {
			nj.Run(ctx)
			wait <- struct{}{}
		}()
		mock.clock.Add(time.Minute)
		cancel()
		<-wait
		require.NoError(t, mock.logger.Error(), "error logged")
	})
	t.Run("notify every event", func(t *testing.T) {
		t.Parallel()
		mock, finish := setup(t)
		defer finish()
		nextInvocation := mock.clock.Now().Add(100 * time.Millisecond)
		mock.repo.EXPECT().GetNearest(gomock.Any()).Return(domains.Event{NextSendTime: nextInvocation}, nil)
		events := []domains.Event{{ID: 0}, {ID: 1}}
		mock.repo.EXPECT().ListNotSended(gomock.Any(), nextInvocation).Return(events, nil)
		for _, ev := range events {
			mock.notifier.EXPECT().Notify(gomock.Any(), domains.NewSendingEvent(ev)).Return(nil)
			mock.repo.EXPECT().Update(gomock.Any(), ev.Rescheule(nextInvocation)).Return(nil)
		}
		mock.repo.EXPECT().GetNearest(gomock.Any()).Return(domains.Event{}, errEventNotFound)
		ctx := context.Background()
		ctx = log.InCtx(ctx, slog.New(mock.logger))
		nj := notifierjob.New(mock.repo, notifierjob.Config{CheckTasksPeriod: time.Minute}, testutils.MockTXManager{}, mock.clock)
		nj.SetNotifier(mock.notifier)
		ctx, cancel := context.WithTimeout(ctx, testTimeout)
		wait := make(chan struct{})
		go func() {
			nj.Run(ctx)
			wait <- struct{}{}
		}()
		mock.clock.Add(time.Minute)
		cancel()
		<-wait
		require.NoError(t, mock.logger.Error(), "error logged")
	})
	t.Run("invocate if there is no events", func(t *testing.T) {
		t.Parallel()
		mock, finish := setup(t)
		defer finish()
		mock.repo.EXPECT().GetNearest(gomock.Any()).Return(domains.Event{}, errEventNotFound)
		mock.repo.EXPECT().ListNotSended(gomock.Any(), gomock.Any()).Return([]domains.Event{}, nil)
		mock.repo.EXPECT().GetNearest(gomock.Any()).Return(domains.Event{}, errEventNotFound)
		ctx := log.InCtx(context.Background(), slog.New(mock.logger))
		nj := notifierjob.New(mock.repo, notifierjob.Config{CheckTasksPeriod: time.Minute}, testutils.MockTXManager{}, mock.clock)
		nj.SetNotifier(mock.notifier)
		ctx, cancel := context.WithTimeout(ctx, testTimeout)
		wait := make(chan struct{})
		go func() {
			nj.Run(ctx)
			wait <- struct{}{}
		}()
		mock.clock.Add(90 * time.Second)
		cancel()
		<-wait
		require.NoError(t, mock.logger.Error(), "error logged")
	})
}
