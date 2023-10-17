package notifier_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/internal/notifier"
	"github.com/Dyleme/Notifier/internal/notifier/mocks"
	"github.com/Dyleme/Notifier/pkg/utils"
)

var eventIDSeq = utils.NewIntSequence()

func newNotif(t time.Time, per time.Duration) domains.SendingNotification {
	return domains.SendingNotification{
		EventID:     eventIDSeq.Next(),
		EventType:   domains.BasicEventType,
		UserID:      0,
		Message:     "",
		Description: "",
		Params: domains.NotificationParams{
			Period: per,
			Params: domains.Params{}, //nolint:exhaustruct // test
		},
		NotificationTime: t,
	}
}

func TestService_RunJob(t *testing.T) {
	t.Parallel()

	t.Run("without notifications", func(t *testing.T) {
		t.Parallel()

		ctx, _ := context.WithTimeout(context.Background(), 100*time.Millisecond) //nolint:govet // test
		ctrl := gomock.NewController(t)
		mockNotifier := mocks.NewMockNotifier(ctrl)
		notifier.New(ctx, mockNotifier, notifier.Config{Period: 10 * time.Millisecond})
		<-ctx.Done()
	})

	t.Run("one notification after period time", func(t *testing.T) {
		t.Parallel()

		ctx, _ := context.WithTimeout(context.Background(), 100*time.Millisecond) //nolint:govet // test
		ctrl := gomock.NewController(t)
		mockNotifier := mocks.NewMockNotifier(ctrl)
		n := notifier.New(ctx, mockNotifier, notifier.Config{Period: 10 * time.Millisecond})
		err := n.Add(ctx, newNotif(time.Now().Add(time.Second), time.Hour))
		require.NoError(t, err)
		<-ctx.Done()
	})

	t.Run("one notification before period time", func(t *testing.T) {
		t.Parallel()

		ctx, _ := context.WithTimeout(context.Background(), 100*time.Millisecond) //nolint:govet // test
		ctrl := gomock.NewController(t)
		mockNotifier := mocks.NewMockNotifier(ctrl)
		n := notifier.New(ctx, mockNotifier, notifier.Config{Period: time.Hour})
		notif := newNotif(time.Now().Add(10*time.Millisecond), time.Hour)
		err := n.Add(ctx, notif)
		mockNotifier.EXPECT().Notify(ctx, notif)
		require.NoError(t, err)
		<-ctx.Done()
	})

	t.Run("retry notification", func(t *testing.T) {
		t.Parallel()

		ctx, _ := context.WithTimeout(context.Background(), 100*time.Millisecond) //nolint:govet // testing
		ctrl := gomock.NewController(t)
		mockNotifier := mocks.NewMockNotifier(ctrl)
		n := notifier.New(ctx, mockNotifier, notifier.Config{Period: time.Hour})

		notif := newNotif(time.Now().Add(10*time.Millisecond), 50*time.Millisecond)
		err := n.Add(ctx, notif)
		mockNotifier.EXPECT().Notify(ctx, notif)
		mockNotifier.EXPECT().Notify(ctx, notif)
		require.NoError(t, err)
		<-ctx.Done()
	})

	t.Run("two notifications at the same time", func(t *testing.T) {
		t.Parallel()

		ctx, _ := context.WithTimeout(context.Background(), 100*time.Millisecond) //nolint:govet // testing
		ctrl := gomock.NewController(t)
		mockNotifier := mocks.NewMockNotifier(ctrl)
		n := notifier.New(ctx, mockNotifier, notifier.Config{Period: time.Hour})

		time.Sleep(20 * time.Millisecond)
		nTime := time.Now().Add(20 * time.Millisecond)
		notif1 := newNotif(nTime, time.Hour)
		notif2 := newNotif(nTime, time.Hour)
		err := n.Add(ctx, notif1)
		require.NoError(t, err)
		err = n.Add(ctx, notif2)
		require.NoError(t, err)
		mockNotifier.EXPECT().Notify(ctx, notif1)
		mockNotifier.EXPECT().Notify(ctx, notif2)
		<-ctx.Done()
	})
}
