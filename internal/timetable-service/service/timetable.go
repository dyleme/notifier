package service

import (
	"context"
	"time"

	"github.com/Dyleme/Notifier/internal/timetable-service/models"
)

func (s *Service) AddTaskToTimetable(ctx context.Context, userID, taskID int, start time.Time, description string) (models.TimetableTask, error) {
	var tt models.TimetableTask

	err := s.repo.Atomic(ctx, func(ctx context.Context, r Repository) error {
		task, err := r.GetTask(ctx, taskID, userID)
		if err != nil {
			return err
		}

		tt, err = r.AddTimetableTask(ctx, task.ToTimetableTask(start, description))
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

func (s *Service) GetTimetableTask(ctx context.Context, userID, timetableTaskID int) (models.TimetableTask, error) {
	tt, err := s.repo.GetTimetableTask(ctx, timetableTaskID, userID)
	if err != nil {
		logError(ctx, err)
		return models.TimetableTask{}, err
	}

	return tt, nil
}

func (s *Service) ListTimetableTasks(ctx context.Context, userID int) ([]models.TimetableTask, error) {
	tts, err := s.repo.ListTimetableTasks(ctx, userID)
	if err != nil {
		logError(ctx, err)
		return nil, err
	}

	return tts, nil
}

func (s *Service) ListTimetableTasksInPeriod(ctx context.Context, userID int, from, to time.Time) ([]models.TimetableTask, error) {
	tts, err := s.repo.ListTimetableTasksInPeriod(ctx, userID, from, to)
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
		tt, err := s.repo.GetTimetableTask(ctx, params.ID, params.UserID)
		if err != nil {
			return err
		}

		task, err := s.repo.GetTask(ctx, tt.TaskID, tt.UserID)
		if err != nil {
			return err
		}

		tt.Description = params.Description
		tt.Done = params.Done
		tt.Start = params.Start
		tt.Finish = params.Start.Add(task.RequiredTime)

		res, err = s.repo.UpdateTimetableTask(ctx, tt)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		logError(ctx, err)
		return models.TimetableTask{}, err
	}

	return res, nil
}

func (s *Service) DeleteTimetableTask(ctx context.Context, userID, timeTableTaskID int) error {
	err := s.repo.DeleteTimetableTask(ctx, timeTableTaskID, userID)
	if err != nil {
		logError(ctx, err)
		return err
	}

	return nil
}
