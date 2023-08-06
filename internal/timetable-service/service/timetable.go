package service

import (
	"context"
	"time"

	"github.com/Dyleme/Notifier/internal/lib/serverrors"
	"github.com/Dyleme/Notifier/internal/timetable-service/models"
)

type TimetableTaskRepository interface {
	Add(context.Context, models.TimetableTask) (models.TimetableTask, error)
	List(ctx context.Context, userID int) ([]models.TimetableTask, error)
	Update(ctx context.Context, timetableTask models.TimetableTask) (models.TimetableTask, error)
	Delete(ctx context.Context, timetableTaskID, userID int) error
	ListInPeriod(ctx context.Context, userID int, from, to time.Time) ([]models.TimetableTask, error)
	Get(ctx context.Context, timetableTaskID, userID int) (models.TimetableTask, error)
	GetNotNotified(ctx context.Context) ([]models.TimetableTask, error)
	MarkNotified(ctx context.Context, ids []int) error
	UpdateNotificationParams(ctx context.Context, timetableTaskID, userID int, params models.NotificationParams) (models.NotificationParams, error)
}

func (s *Service) CreateTimetableTask(ctx context.Context, task models.Task, timetableTask models.TimetableTask) (models.TimetableTask, error) {
	var tt models.TimetableTask

	err := s.repo.Atomic(ctx, func(ctx context.Context, r Repository) error {
		var err error
		task, err = r.Tasks().Add(ctx, task.UsedTask())
		if err != nil {
			return err
		}

		timetableTask.TaskID = task.ID

		tt, err = r.TimetableTasks().Add(ctx, timetableTask)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logError(ctx, err)
		return models.TimetableTask{}, err
	}

	return tt, nil
}

func (s *Service) AddTaskToTimetable(ctx context.Context, userID, taskID int, start time.Time, description string) (models.TimetableTask, error) {
	var tt models.TimetableTask

	err := s.repo.Atomic(ctx, func(ctx context.Context, r Repository) error {
		task, err := r.Tasks().Get(ctx, taskID, userID)
		if err != nil {
			return err
		}
		if !task.CanUse() {
			return serverrors.NewBusinessLogicError("task is already used")
		}

		ttTask := task.ToTimetableTask(start, description)
		tt, err = r.TimetableTasks().Add(ctx, ttTask)
		if err != nil {
			return err
		}

		updatedTask := task.UsedTask()
		err = r.Tasks().Update(ctx, updatedTask)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return models.TimetableTask{}, err
	}

	return tt, nil
}

func (s *Service) GetTimetableTask(ctx context.Context, userID, timetableTaskID int) (models.TimetableTask, error) {
	tt, err := s.repo.TimetableTasks().Get(ctx, timetableTaskID, userID)
	if err != nil {
		logError(ctx, err)
		return models.TimetableTask{}, err
	}

	return tt, nil
}

func (s *Service) ListTimetableTasks(ctx context.Context, userID int) ([]models.TimetableTask, error) {
	tts, err := s.repo.TimetableTasks().List(ctx, userID)
	if err != nil {
		logError(ctx, err)
		return nil, err
	}

	return tts, nil
}

func (s *Service) ListTimetableTasksInPeriod(ctx context.Context, userID int, from, to time.Time) ([]models.TimetableTask, error) {
	tts, err := s.repo.TimetableTasks().ListInPeriod(ctx, userID, from, to)
	if err != nil {
		logError(ctx, err)
		return nil, err
	}

	return tts, nil
}

type UpdateTimetableParams struct {
	ID          int
	UserID      int
	Description string
	Start       time.Time
	Done        bool
}

func (s *Service) UpdateTimetable(ctx context.Context, params UpdateTimetableParams) (models.TimetableTask, error) {
	var res models.TimetableTask
	err := s.repo.Atomic(ctx, func(ctx context.Context, repo Repository) error {
		tt, err := repo.TimetableTasks().Get(ctx, params.ID, params.UserID)
		if err != nil {
			return err
		}

		task, err := repo.Tasks().Get(ctx, tt.TaskID, tt.UserID)
		if err != nil {
			return err
		}

		tt.Description = params.Description
		tt.Done = params.Done
		tt.Start = params.Start
		tt.Finish = params.Start.Add(task.RequiredTime)

		res, err = s.repo.TimetableTasks().Update(ctx, tt)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return models.TimetableTask{}, err
	}

	return res, nil
}

func (s *Service) DeleteTimetableTask(ctx context.Context, userID, timeTableTaskID int) error {
	err := s.repo.TimetableTasks().Delete(ctx, timeTableTaskID, userID)
	if err != nil {
		logError(ctx, err)
		return err
	}

	return nil
}
