package notifier_job_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/avito-tech/go-transaction-manager/trm/manager"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/internal/service/service"
	"github.com/Dyleme/Notifier/internal/service/service/mocks"
	"github.com/Dyleme/Notifier/pkg/log"
	"github.com/Dyleme/Notifier/pkg/log/mocklogger"
	"github.com/Dyleme/Notifier/pkg/serverrors"
	"github.com/Dyleme/Notifier/pkg/testutils"
	"github.com/Dyleme/Notifier/pkg/utils"
)

func setupTest(ctrl *gomock.Controller, testTime time.Duration) (context.Context, *mocksUnion, *mocklogger.MockHandler) {
	mockHandler := mocklogger.NewHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}) //nolint:exhaustruct //testing
	ctx, _ := context.WithTimeout(context.Background(), testTime)                                 //nolint:govet // testing
	ctx = log.InCtx(ctx, slog.New(mockHandler))
	allMocks := mocksUnion{
		events:         mocks.NewMockEventRepository(ctrl),
		periodicEvents: mocks.NewMockPeriodicEventsRepository(ctrl),
		defaultNotif:   mocks.NewMockNotificationParamsRepository(ctrl),
		notifier:       mocks.NewMockNotifier(ctrl),
	}

	return ctx, &allMocks, mockHandler
}

func listEvents(mcks *mocksUnion, basicEvents []domains.Event, periodicEvents []domains.PeriodicEvent) {
	if len(basicEvents) != 0 {
		mcks.events.EXPECT().ListEventsBefore(gomock.Any(), gomock.Any()).Return(basicEvents, nil)
	} else {
		mcks.events.EXPECT().ListEventsBefore(gomock.Any(), gomock.Any()).Return(nil, serverrors.NewNotFoundError(fmt.Errorf("err"), "event"))
	}

	if len(periodicEvents) != 0 {
		mcks.periodicEvents.EXPECT().ListNotificationsAtSendTime(gomock.Any(), gomock.Any()).Return(periodicEvents, nil)
	} else {
		mcks.periodicEvents.EXPECT().ListNotificationsAtSendTime(gomock.Any(), gomock.Any()).Return(nil, serverrors.NewNotFoundError(fmt.Errorf("err"), "periodicEvent"))
	}
}

func getTime(mcks *mocksUnion, basicEventTime, periodicEventTime time.Time) {
	if !basicEventTime.IsZero() {
		mcks.events.EXPECT().GetNearestEventSendTime(gomock.Any()).Return(basicEventTime, nil)
	} else {
		mcks.events.EXPECT().GetNearestEventSendTime(gomock.Any()).Return(time.Time{}, serverrors.NewNotFoundError(fmt.Errorf("err"), "event"))
	}

	if !periodicEventTime.IsZero() {
		mcks.periodicEvents.EXPECT().GetNearestNotificationSendTime(gomock.Any()).Return(periodicEventTime, nil)
	} else {
		mcks.periodicEvents.EXPECT().GetNearestNotificationSendTime(gomock.Any()).Return(time.Time{}, serverrors.NewNotFoundError(fmt.Errorf("err"), "event"))
	}
}

var defaultNotifParams = domains.NotificationParams{ //nolint:exhaustruct //no need to fill
	Period: time.Hour,
}

func sendNotifications(mcks *mocksUnion, basicEvents []domains.Event, periodicEvents []domains.PeriodicEvent) {
	listEvents(mcks, basicEvents, periodicEvents)
	for _, basicEvent := range basicEvents {
		mcks.defaultNotif.EXPECT().Get(gomock.Any(), basicEvent.UserID).Return(defaultNotifParams, nil)
		mcks.notifier.EXPECT().Add(gomock.Any(), gomock.Any())
		mcks.events.EXPECT().MarkNotified(gomock.Any(), basicEvent.ID)
	}
	for _, perEvent := range periodicEvents {
		mcks.defaultNotif.EXPECT().Get(gomock.Any(), perEvent.UserID).Return(defaultNotifParams, nil)
		mcks.notifier.EXPECT().Add(gomock.Any(), gomock.Any())
		mcks.periodicEvents.EXPECT().Get(gomock.Any(), perEvent.ID, perEvent.UserID).Return(perEvent, nil)
		mcks.periodicEvents.EXPECT().MarkNotificationSend(gomock.Any(), perEvent.ID)
	}
}

var eventsSeq = utils.NewSequence(domains.Event{ //nolint:exhaustruct //test
	ID:          1,
	UserID:      1,
	Text:        strconv.Itoa(1),
	Description: strconv.Itoa(1),
}, func(t domains.Event) domains.Event {
	i := t.ID
	i++

	return domains.Event{ //nolint:exhaustruct //test
		ID:          i,
		UserID:      i,
		Text:        strconv.Itoa(i),
		Description: strconv.Itoa(i),
	}
})

var periodicEventSeq = utils.NewSequence(domains.PeriodicEvent{ //nolint:exhaustruct //test
	ID:             1,
	Text:           strconv.Itoa(1),
	Description:    strconv.Itoa(1),
	UserID:         1,
	SmallestPeriod: 1,
	BiggestPeriod:  2,
	Notification: domains.PeriodicEventNotification{ //nolint:exhaustruct //test
		ID:              1,
		PeriodicEventID: 1,
	},
}, func(t domains.PeriodicEvent) domains.PeriodicEvent {
	i := t.ID
	i++

	return domains.PeriodicEvent{ //nolint:exhaustruct //test need to fill
		ID:             i,
		Text:           strconv.Itoa(i),
		Description:    strconv.Itoa(i),
		UserID:         i,
		SmallestPeriod: 1,
		BiggestPeriod:  2,
		Notification: domains.PeriodicEventNotification{ //nolint:exhaustruct //test
			ID:              i,
			PeriodicEventID: i,
		},
	}
})

type mocksUnion struct {
	events         *mocks.MockEventRepository
	periodicEvents *mocks.MockPeriodicEventsRepository
	defaultNotif   *mocks.MockNotificationParamsRepository
	notifier       *mocks.MockNotifier
}

func Test_notifierJob_RunJob(t *testing.T) {
	t.Parallel()

	t.Run("no notifications", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		testTime := 120 * time.Millisecond
		checkTaskPeriod := 50 * time.Millisecond
		ctx, serviceMocks, mockLoggerHandler := setupTest(ctrl, testTime)

		for i := 0; i < 3; i++ {
			sendNotifications(serviceMocks, nil, nil)
			getTime(serviceMocks, time.Time{}, time.Time{})
		}

		repo := &RepositoryMock{ //nolint:exhaustruct //tests
			DefaultNotificationRepo: serviceMocks.defaultNotif,
			EventsRepo:              serviceMocks.events,
			PeriodicEventsRepo:      serviceMocks.periodicEvents,
		}
		nj := service.NewNotifierJob(repo, serviceMocks.notifier, service.Config{CheckTasksPeriod: checkTaskPeriod}, manager.Must(testutils.TxManager))
		nj.RunJob(ctx)

		require.NoError(t, mockLoggerHandler.Error())
	})

	t.Run("only basic events", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		testTime := 120 * time.Millisecond
		checkTaskPeriod := 1000 * time.Millisecond
		ctx, serviceMocks, mockLoggerHandler := setupTest(ctrl, testTime)

		events := eventsSeq.Generate(2)
		sendNotifications(serviceMocks, events, nil)
		getTime(serviceMocks, time.Now().Add(20*time.Millisecond), time.Time{})
		newEvents := eventsSeq.Generate(2)
		sendNotifications(serviceMocks, newEvents, nil)
		getTime(serviceMocks, time.Time{}, time.Time{})
		repo := &RepositoryMock{ //nolint:exhaustruct //tests
			DefaultNotificationRepo: serviceMocks.defaultNotif,
			EventsRepo:              serviceMocks.events,
			PeriodicEventsRepo:      serviceMocks.periodicEvents,
		}
		nj := service.NewNotifierJob(repo, serviceMocks.notifier, service.Config{CheckTasksPeriod: checkTaskPeriod}, manager.Must(testutils.TxManager))
		nj.RunJob(ctx)

		require.NoError(t, mockLoggerHandler.Error())
	})

	t.Run("only periodic events", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		testTime := 120 * time.Millisecond
		checkTaskPeriod := 1000 * time.Millisecond
		ctx, serviceMocks, mockLoggerHandler := setupTest(ctrl, testTime)

		events := periodicEventSeq.Generate(2)
		sendNotifications(serviceMocks, nil, events)
		getTime(serviceMocks, time.Time{}, time.Now().Add(20*time.Millisecond))
		newEvents := periodicEventSeq.Generate(2)
		sendNotifications(serviceMocks, nil, newEvents)
		getTime(serviceMocks, time.Time{}, time.Time{})
		repo := &RepositoryMock{ //nolint:exhaustruct //tests
			DefaultNotificationRepo: serviceMocks.defaultNotif,
			EventsRepo:              serviceMocks.events,
			PeriodicEventsRepo:      serviceMocks.periodicEvents,
		}
		nj := service.NewNotifierJob(repo, serviceMocks.notifier, service.Config{CheckTasksPeriod: checkTaskPeriod}, manager.Must(testutils.TxManager))
		nj.RunJob(ctx)

		require.NoError(t, mockLoggerHandler.Error())
	})
}

func TestNotifierJob_UpdateWithTime(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		period   time.Duration
		isUpdate bool
	}{
		{
			name:     "no update",
			period:   2 * time.Hour,
			isUpdate: false,
		},
		{
			name:     "with update",
			period:   20 * time.Millisecond,
			isUpdate: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			testTime := 180 * time.Millisecond
			checkPeriod := time.Hour
			waitTime := 10 * time.Millisecond
			ctx, serviceMocks, mockHandlerLogger := setupTest(ctrl, testTime)
			sendNotifications(serviceMocks, nil, nil)
			getTime(serviceMocks, time.Time{}, time.Time{})
			if tc.isUpdate {
				getTime(serviceMocks, time.Now().Add(waitTime+tc.period), time.Time{})
				sendNotifications(serviceMocks, nil, nil)
				getTime(serviceMocks, time.Time{}, time.Time{})
			}
			repo := &RepositoryMock{ //nolint:exhaustruct // mock
				DefaultNotificationRepo: serviceMocks.defaultNotif,
				EventsRepo:              serviceMocks.events,
				PeriodicEventsRepo:      serviceMocks.periodicEvents,
			}
			nj := service.NewNotifierJob(repo, serviceMocks.notifier, service.Config{CheckTasksPeriod: checkPeriod}, manager.Must(testutils.TxManager))
			wait := make(chan struct{})
			go func() {
				nj.RunJob(ctx)
				wait <- struct{}{}
			}()
			time.Sleep(waitTime)
			nj.UpdateWithTime(ctx, time.Now().Add(tc.period))
			<-wait

			require.NoError(t, mockHandlerLogger.Error())
		})
	}
}
