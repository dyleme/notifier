package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/pkg/serverrors"
)

//go:generate mockgen -destination=mocks/tasks_mocks.go -package=mocks . BasicTaskRepository
type BasicTaskRepository interface {
	Add(ctx context.Context, task domains.BasicTask) (domains.BasicTask, error)
	List(ctx context.Context, userID int, params ListParams) ([]domains.BasicTask, error)
	Update(ctx context.Context, task domains.BasicTask) (domains.BasicTask, error)
	Delete(ctx context.Context, taskID int) error
	Get(ctx context.Context, taskID int) (domains.BasicTask, error)
}

func (s *Service) CreateTask(ctx context.Context, task domains.BasicTask) (domains.BasicTask, error) {
	op := "Service.CreateTask: %w"
	var createdTask domains.BasicTask

	err := s.tr.Do(ctx, func(ctx context.Context) error {
		var err error

		createdTask, err = s.repo.Tasks().Add(ctx, task)
		if err != nil {
			return fmt.Errorf("add task: %w", err)
		}

		notif := createdTask.NewNotification()
		_, err = s.repo.Notifications().Add(ctx, notif)
		if err != nil {
			return fmt.Errorf("add notification: %w", err)
		}

		return nil
	})
	if err != nil {
		err = fmt.Errorf(op, err)
		logError(ctx, err)

		return domains.BasicTask{}, err
	}

	s.notifierJob.UpdateWithTime(ctx, createdTask.Start)

	return createdTask, nil
}

func (s *Service) GetTask(ctx context.Context, userID, taskID int) (domains.BasicTask, error) {
	op := "Service.GetTask: %w"
	tt, err := s.repo.Tasks().Get(ctx, taskID)
	if err != nil {
		err = fmt.Errorf(op, err)
		logError(ctx, err)

		return domains.BasicTask{}, err
	}

	if !tt.BelongsTo(userID) {
		return domains.BasicTask{}, fmt.Errorf("belongs to: %w", serverrors.NewNotFoundError(err, "task"))
	}

	return tt, nil
}

func (s *Service) ListTasks(ctx context.Context, userID int, listParams ListParams) ([]domains.BasicTask, error) {
	op := "Service.ListTasks: %w"
	tts, err := s.repo.Tasks().List(ctx, userID, listParams)
	if err != nil {
		err = fmt.Errorf(op, err)
		logError(ctx, err)

		return nil, err
	}

	return tts, nil
}

func (s *Service) UpdateBasicTask(ctx context.Context, params domains.BasicTask, userID int) (domains.BasicTask, error) {
	var task domains.BasicTask
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		t, err := s.repo.Tasks().Get(ctx, params.ID)
		if err != nil {
			return fmt.Errorf("get task: %w", err)
		}

		if !t.BelongsTo(userID) {
			return serverrors.NewNotFoundError(err, "task")
		}

		t.Text = params.Text
		t.Description = params.Description
		t.Start = params.Start

		task, err = s.repo.Tasks().Update(ctx, t)
		if err != nil {
			return fmt.Errorf("update task: %w", err)
		}

		notif, err := s.repo.Notifications().GetLatest(ctx, t.ID)
		if err != nil {
			return fmt.Errorf("get latest notification: %w", err)
		}

		notif.Text = params.Text
		notif.Description = params.Description
		notif.SendTime = params.Start

		err = s.repo.Notifications().Update(ctx, notif)
		if err != nil {
			return fmt.Errorf("update notification: %w", err)
		}

		return nil
	})
	if err != nil {
		err = fmt.Errorf("tr: %w", err)
		logError(ctx, err)

		return domains.BasicTask{}, err
	}

	s.notifierJob.UpdateWithTime(ctx, task.Start)

	return task, nil
}

func (s *Service) createNewNotificationForPeriodicTask(ctx context.Context, taskID, userID int) error {
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		bt, err := s.repo.PeriodicTasks().Get(ctx, taskID)
		if err != nil {
			return fmt.Errorf("periodic tasks get[taskID=%v,userID=%v]: %w", taskID, userID, err)
		}
		if !bt.BelongsTo(userID) {
			return serverrors.NewBusinessLogicError("task does not belong to user")
		}

		nextNotif, err := bt.NewNotification(time.Now())
		if err != nil {
			return fmt.Errorf("new notification: %w", serverrors.NewBusinessLogicError(err.Error()))
		}

		_, err = s.repo.Notifications().Add(ctx, nextNotif)
		if err != nil {
			return fmt.Errorf("periodic tasks add notification: %w", err)
		}

		s.notifierJob.UpdateWithTime(ctx, nextNotif.SendTime)

		return nil
	})
	if err != nil {
		return fmt.Errorf("tr: %w", err)
	}

	return nil
}

func (s *Service) SetNotificationDoneStatus(ctx context.Context, notifID, userID int, done bool) error {
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		notif, err := s.repo.Notifications().Get(ctx, notifID)
		if err != nil {
			return fmt.Errorf("get notification: %w", err)
		}
		notif.Done = done
		err = s.repo.Notifications().Update(ctx, notif)
		if err != nil {
			return fmt.Errorf("update notification: %w", err)
		}
		err = s.notifier.Delete(ctx, notif.ID)
		if err != nil {
			return fmt.Errorf("delete notifier notification: %w", err)
		}
		switch notif.TaskType {
		case domains.BasicTaskType:
		case domains.PeriodicTaskType:
			err := s.createNewNotificationForPeriodicTask(ctx, notif.TaskID, userID)
			if err != nil {
				return fmt.Errorf("setTaskDoneStatusPeriodicTask: %w", err)
			}
		default:
			return fmt.Errorf("unknown taskType[%v]", notif.TaskType)
		}

		err = s.notifier.Delete(ctx, notifID)
		if err != nil {
			return fmt.Errorf("delete[notifID=%v]: %w", notifID, err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("tr: %w", err)
	}

	return nil
}

func (s *Service) DeleteBasicTask(ctx context.Context, userID, taskID int) error {
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		task, err := s.repo.Tasks().Get(ctx, taskID)
		if err != nil {
			return fmt.Errorf("get task: %w", err)
		}

		if !task.BelongsTo(userID) {
			return serverrors.NewNotFoundError(err, "basic task")
		}

		notif, err := s.repo.Notifications().GetLatest(ctx, taskID)
		if err != nil {
			return fmt.Errorf("get latest notification: %w", err)
		}

		err = s.repo.Notifications().Delete(ctx, notif.ID)
		if err != nil {
			return fmt.Errorf("delete notification: %w", err)
		}

		err = s.repo.Tasks().Delete(ctx, taskID)
		if err != nil {
			return fmt.Errorf("delete basic task: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("tr: %w", err)
	}

	return nil
}
