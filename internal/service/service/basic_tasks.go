package service

import (
	"context"
	"fmt"

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/pkg/log"
	"github.com/Dyleme/Notifier/pkg/serverrors"
)

//go:generate mockgen -destination=mocks/basic_tasks_mocks.go -package=mocks . BasicTaskRepository
type BasicTaskRepository interface {
	Add(ctx context.Context, task domains.BasicTask) (domains.BasicTask, error)
	List(ctx context.Context, userID int, params ListParams) ([]domains.BasicTask, error)
	Update(ctx context.Context, task domains.BasicTask) (domains.BasicTask, error)
	Delete(ctx context.Context, taskID int) error
	Get(ctx context.Context, taskID int) (domains.BasicTask, error)
}

func (s *Service) CreateBasicTask(ctx context.Context, task domains.BasicTask) (domains.BasicTask, error) {
	op := "Service.CreateTask: %w"
	var createdTask domains.BasicTask
	var createdEvent domains.Event

	err := s.tr.Do(ctx, func(ctx context.Context) error {
		var err error

		log.Ctx(ctx).Debug("add task", "event", task)
		createdTask, err = s.repos.basicTasks.Add(ctx, task)
		if err != nil {
			return fmt.Errorf("add task: %w", err)
		}

		err = s.createAndAddEvent(ctx, createdTask, task.UserID)
		if err != nil {
			return fmt.Errorf("create and add event: %w", err)
		}

		return nil
	})
	if err != nil {
		err = fmt.Errorf(op, err)
		logError(ctx, err)

		return domains.BasicTask{}, err
	}

	s.notifierJob.UpdateWithTime(ctx, createdEvent.NextSendTime)

	return createdTask, nil
}

func (s *Service) GetBasicTask(ctx context.Context, userID, taskID int) (domains.BasicTask, error) {
	op := "Service.GetTask: %w"
	tt, err := s.repos.basicTasks.Get(ctx, taskID)
	if err != nil {
		err = fmt.Errorf(op, err)
		logError(ctx, err)

		return domains.BasicTask{}, err
	}

	if err := tt.BelongsTo(userID); err != nil {
		return domains.BasicTask{}, fmt.Errorf("belongs to: %w", serverrors.NewBusinessLogicError(err.Error()))
	}

	return tt, nil
}

func (s *Service) ListBasicTasks(ctx context.Context, userID int, listParams ListParams) ([]domains.BasicTask, error) {
	op := "Service.ListTasks: %w"
	tts, err := s.repos.basicTasks.List(ctx, userID, listParams)
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
		t, err := s.repos.basicTasks.Get(ctx, params.ID)
		if err != nil {
			return fmt.Errorf("get task: %w", err)
		}

		if err := t.BelongsTo(userID); err != nil {
			return fmt.Errorf("belongs to: %w", serverrors.NewBusinessLogicError(err.Error()))
		}

		t.Text = params.Text
		t.Description = params.Description
		t.Start = params.Start

		task, err = s.repos.basicTasks.Update(ctx, t)
		if err != nil {
			return fmt.Errorf("update task: %w", err)
		}

		event, err := s.repos.events.GetLatest(ctx, t.ID, domains.BasicTaskType)
		if err != nil {
			return fmt.Errorf("get latest event: %w", err)
		}

		updatedEvent, err := task.UpdatedEvent(event)
		if err != nil {
			return fmt.Errorf("update event: %w", err)
		}

		err = s.repos.events.Update(ctx, updatedEvent)
		if err != nil {
			return fmt.Errorf("update event: %w", err)
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

func (s *Service) DeleteBasicTask(ctx context.Context, userID, taskID int) error {
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		task, err := s.repos.basicTasks.Get(ctx, taskID)
		if err != nil {
			return fmt.Errorf("get task: %w", err)
		}

		if err := task.BelongsTo(userID); err != nil {
			return fmt.Errorf("belongs to: %w", serverrors.NewBusinessLogicError(err.Error()))
		}

		event, err := s.repos.events.GetLatest(ctx, taskID, domains.BasicTaskType)
		if err != nil {
			return fmt.Errorf("get latest event: %w", err)
		}

		err = s.repos.events.Delete(ctx, event.ID)
		if err != nil {
			return fmt.Errorf("delete event: %w", err)
		}

		err = s.repos.basicTasks.Delete(ctx, taskID)
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
