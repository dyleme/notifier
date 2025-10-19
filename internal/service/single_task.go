package service

import (
	"context"
	"fmt"

	"github.com/dyleme/Notifier/internal/domain"
	"github.com/dyleme/Notifier/pkg/log"
	"github.com/dyleme/Notifier/pkg/utils/slice"
)

func (s *Service) CreateSingleTask(ctx context.Context, singleTask domain.SingleTask) error {
	log.Ctx(ctx).Debug("creating single task", "single task", singleTask)

	createdEvent := singleTask.NewEvent()

	err := s.tr.Do(ctx, func(ctx context.Context) error {
		return s.addTask(ctx, singleTask.BuildTask(), createdEvent)
	})
	if err != nil {
		return fmt.Errorf("tr: %w", err)
	}

	s.notifierJob.UpdateWithTime(ctx, createdEvent.OriginalSending)

	return nil
}

func (s *Service) GetSingleTask(ctx context.Context, userID, taskID int) (domain.SingleTask, error) {
	log.Ctx(ctx).Debug("getting single task", "taskID", taskID)
	task, err := s.repos.tasks.Get(ctx, taskID, userID)
	if err != nil {
		return domain.SingleTask{}, fmt.Errorf("get basic task userID[%v], taskID[%v]: %w", userID, taskID, err)
	}

	return domain.ParseSingleTask(task)
}

func (s *Service) ListSingleTasks(ctx context.Context, userID int, params ListParams) ([]domain.SingleTask, error) {
	tasks, err := s.repos.tasks.List(ctx, userID, domain.Single, params)
	if err != nil {
		return nil, fmt.Errorf("list tasks userID[%v]: %w", userID, err)
	}

	return slice.DtoError(tasks, domain.ParseSingleTask)
}

func (s *Service) UpdateSingleTask(ctx context.Context, singleTask domain.SingleTask, userID int) error {
	log.Ctx(ctx).Debug("updating single task", "task", singleTask, "userID", userID)
	updatedEvent := singleTask.NewEvent()
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		err := s.updateTask(ctx, singleTask.BuildTask(), updatedEvent)
		if err != nil {
			return fmt.Errorf("update task: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("tr: %w", err)
	}

	s.notifierJob.UpdateWithTime(ctx, updatedEvent.OriginalSending)

	return nil
}
