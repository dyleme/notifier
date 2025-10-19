package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/dyleme/Notifier/internal/domain"
	"github.com/dyleme/Notifier/pkg/log"
	"github.com/dyleme/Notifier/pkg/utils/slice"
)

func (s *Service) CreatePeriodicTask(ctx context.Context, perTask domain.PeriodicTask) error {
	log.Ctx(ctx).Debug("creating periodic task", slog.Any("periodic task", perTask))

	createdEvent := perTask.NewSending(time.Now())

	err := s.tr.Do(ctx, func(ctx context.Context) error {
		return s.addTask(ctx, perTask.BuildTask(), createdEvent)
	})
	if err != nil {
		return fmt.Errorf("tr: %w", err)
	}

	return nil
}

func (s *Service) GetPeriodicTask(ctx context.Context, taskID, userID int) (domain.PeriodicTask, error) {
	log.Ctx(ctx).Debug("getting periodic task", "taskID", taskID)
	task, err := s.repos.tasks.Get(ctx, taskID, userID)
	if err != nil {
		return domain.PeriodicTask{}, fmt.Errorf("get[taskID=%v]: %w", taskID, err)
	}

	return domain.ParsePeriodicTask(task)
}

func (s *Service) ListPeriodicTasks(ctx context.Context, userID int, params ListParams) ([]domain.PeriodicTask, error) {
	tasks, err := s.repos.tasks.List(ctx, userID, domain.Periodic, params)
	if err != nil {
		return nil, fmt.Errorf("list tasks userID[%v]: %w", userID, err)
	}

	return slice.DtoError(tasks, domain.ParsePeriodicTask)
}

func (s *Service) UpdatePeriodicTask(ctx context.Context, perTask domain.PeriodicTask) error {
	updatedEvent := perTask.NewSending(time.Now())
	err := s.tr.Do(ctx, func(ctx context.Context) error {
		err := s.updateTask(ctx, perTask.BuildTask(), updatedEvent)
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
