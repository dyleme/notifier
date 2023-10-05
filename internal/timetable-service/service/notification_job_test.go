package service_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/Dyleme/Notifier/internal/lib/log"
	"github.com/Dyleme/Notifier/internal/lib/log/mocklogger"
	"github.com/Dyleme/Notifier/internal/lib/serverrors"
	"github.com/Dyleme/Notifier/internal/lib/utils/ptr"
	"github.com/Dyleme/Notifier/internal/timetable-service/domains"
	"github.com/Dyleme/Notifier/internal/timetable-service/service"
	"github.com/Dyleme/Notifier/internal/timetable-service/service/mocks"
)

func Test_notifierJob_RunJob(t *testing.T) {
	t.Parallel()

	t.Run("all times", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		testTime := 180 * time.Millisecond
		ctx, _ := context.WithTimeout(context.Background(), testTime)                                //nolint:govet // testing
		mockLogger := mocklogger.NewHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}) //nolint:exhaustruct //testing
		ctx = log.InCtx(ctx, slog.New(mockLogger))
		eventsRepo := mocks.NewMockEventRepository(ctrl)
		timeToNearestCall := 20 * time.Millisecond
		nearestCallTime := time.Now().Add(timeToNearestCall)
		checkTaskPeriod := 50 * time.Millisecond
		defaultNotifParamsRepo := mocks.NewMockNotificationParamsRepository(ctrl)
		notifier := mocks.NewMockNotifier(ctrl)
		events := []domains.Event{
			{ //nolint:exhaustruct //tests
				ID:     1,
				UserID: 1,
			},
			{ //nolint:exhaustruct //tests
				ID:     2,
				UserID: 2,
			},
		}
		eventsRepo.EXPECT().GetNearestEventSendTime(ctx).Return(nearestCallTime, nil)
		notifyEvents(ctx, eventsRepo, defaultNotifParamsRepo, notifier, events, nearestCallTime)
		eventsRepo.EXPECT().GetNearestEventSendTime(ctx).Return(time.Time{}, serverrors.NewNotFoundError(fmt.Errorf("err"), "nearest time"))
		notifyEventsJobCalls(ctx, int((testTime-timeToNearestCall)/checkTaskPeriod), eventsRepo)
		repo := &RepositoryMock{ //nolint:exhaustruct //tests
			DefaultNotificationRepo: defaultNotifParamsRepo,
			EventsRepo:              eventsRepo,
		}
		nj := service.NewNotifierJob(repo, notifier, service.Config{CheckTasksPeriod: checkTaskPeriod})
		nj.RunJob(ctx)

		assert.NoError(t, mockLogger.Error())
	})
}

type callParams struct {
	events         []domains.Event
	timeToNextCall *time.Duration
}

func anyNotifyEvents(ctx context.Context, eventsRepo *mocks.MockEventRepository, defaultNotifParamsRepo *mocks.MockNotificationParamsRepository, notifier *mocks.MockNotifier, params []callParams, nearestCallTime time.Time) {
	eventsRepo.EXPECT().GetNearestEventSendTime(ctx).Return(nearestCallTime, nil)
	for i, p := range params {
		if i != len(params)-1 {
			if len(p.events) != 0 {
				eventsRepo.EXPECT().ListEventsAtSendTime(ctx, gomock.Any()).Return(p.events, nil)
				for _, ev := range p.events {
					defaultNotifParamsRepo.EXPECT().Get(ctx, ev.UserID)
					eventsRepo.EXPECT().MarkNotified(ctx, ev.ID)
					notifier.EXPECT().Add(ctx, gomock.Any())
				}
			} else {
				eventsRepo.EXPECT().ListEventsAtSendTime(ctx, gomock.Any()).Return(nil, serverrors.NewNotFoundError(fmt.Errorf("not found"), "nearest time"))
			}
		}
		if p.timeToNextCall != nil {
			eventsRepo.EXPECT().GetNearestEventSendTime(ctx).Return(time.Now().Add(*p.timeToNextCall), nil)
		} else {
			eventsRepo.EXPECT().GetNearestEventSendTime(ctx).Return(time.Time{}, serverrors.NewNotFoundError(fmt.Errorf("err"), "nearest time"))
		}
	}
}

func notifyEvents(ctx context.Context, eventsRepo *mocks.MockEventRepository, defaultNotifParamsRepo *mocks.MockNotificationParamsRepository, notifier *mocks.MockNotifier, events []domains.Event, nearestCallTime time.Time) {
	eventsRepo.EXPECT().ListEventsAtSendTime(ctx, nearestCallTime).Return(events, nil)
	for _, ev := range events {
		defaultNotifParamsRepo.EXPECT().Get(ctx, ev.UserID)
		eventsRepo.EXPECT().MarkNotified(ctx, ev.ID)
		notifier.EXPECT().Add(ctx, gomock.Any())
	}
}

func notifyEventsJobCalls(ctx context.Context, noEventsCalls int, eventsRepo *mocks.MockEventRepository) {
	for i := 0; i < noEventsCalls; i++ {
		eventsRepo.EXPECT().ListEventsAtSendTime(ctx, gomock.Any()).Return([]domains.Event{}, nil)
		eventsRepo.EXPECT().GetNearestEventSendTime(ctx).Return(time.Time{}, serverrors.NewNotFoundError(fmt.Errorf("err"), "nearest time"))
	}
}

func TestNotifierJob_UpdateWithTime(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		period time.Duration
		params []callParams
	}{
		// {
		// 	name:   "no update",
		// 	period: time.Hour,
		// 	params: []callParams{
		// 		{
		// 			events: []domains.Service{
		// 				{
		// 					ID:     1,
		// 					UserID: 1,
		// 				},
		// 				{
		// 					ID:     2,
		// 					UserID: 2,
		// 				},
		// 			},
		// 			timeToNextCall: ptr.Ptr(80 * time.Millisecond),
		// 		},
		// 		{},
		// 	},
		// },
		{
			name:   "with update",
			period: 20 * time.Millisecond,
			params: []callParams{
				{
					events: []domains.Event{
						{ //nolint:exhaustruct //no need to fill
							ID:     1,
							UserID: 1,
						},
						{ //nolint:exhaustruct //no need to fill
							ID:     2,
							UserID: 2,
						},
					},
					timeToNextCall: ptr.Ptr(20 * time.Millisecond),
				},
				{}, //nolint:exhaustruct //no need to fill
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			testTime := 100 * time.Millisecond
			ctx, _ := context.WithTimeout(context.Background(), testTime)                                //nolint:govet // testing
			mockLogger := mocklogger.NewHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}) //nolint:exhaustruct // mock
			ctx = log.InCtx(ctx, slog.New(mockLogger))
			eventsRepo := mocks.NewMockEventRepository(ctrl)
			defaultNotifParamsRepo := mocks.NewMockNotificationParamsRepository(ctrl)
			notifier := mocks.NewMockNotifier(ctrl)
			timeToNearestCall := 80 * time.Millisecond
			nearestCallTime := time.Now().Add(timeToNearestCall)
			anyNotifyEvents(ctx, eventsRepo, defaultNotifParamsRepo, notifier, tc.params, nearestCallTime)
			repo := &RepositoryMock{ //nolint:exhaustruct // mock
				DefaultNotificationRepo: defaultNotifParamsRepo,
				EventsRepo:              eventsRepo,
			}
			nj := service.NewNotifierJob(repo, notifier, service.Config{CheckTasksPeriod: time.Hour})
			wait := make(chan struct{})
			go func() {
				nj.RunJob(ctx)
				wait <- struct{}{}
			}()
			time.Sleep(10 * time.Millisecond)
			nj.UpdateWithTime(ctx, time.Now().Add(tc.period))
			<-wait

			assert.NoError(t, mockLogger.Error())
		})
	}
}
