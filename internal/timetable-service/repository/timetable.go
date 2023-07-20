package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Dyleme/Notifier/internal/lib/serverrors"
	"github.com/Dyleme/Notifier/internal/lib/sql/pgxconv"
	"github.com/Dyleme/Notifier/internal/timetable-service/models"
	"github.com/Dyleme/Notifier/internal/timetable-service/repository/queries"
)

func dtoTimetableTask(t queries.TimetableTask) models.TimetableTask {
	return models.TimetableTask{
		ID:          int(t.ID),
		UserID:      int(t.UserID),
		TaskID:      int(t.TaskID),
		Text:        t.Text,
		Description: pgxconv.String(t.Description),
		Start:       pgxconv.Time(t.Start),
		Finish:      pgxconv.Time(t.Finish),
		Done:        t.Done,
	}
}

func dtoTimetableTasks(ts []queries.TimetableTask) []models.TimetableTask {
	timetables := make([]models.TimetableTask, 0, len(ts))

	for _, t := range ts { //nolint:gocritic // parsing database range
		timetables = append(timetables, dtoTimetableTask(t))
	}

	return timetables
}

func (r *Repository) AddTimetableTask(ctx context.Context, tt models.TimetableTask) (models.TimetableTask, error) {
	op := "add timetable task: %w"
	addedTimetable, err := r.q.AddTimetableTask(ctx, queries.AddTimetableTaskParams{
		UserID:      int32(tt.UserID),
		TaskID:      int32(tt.TaskID),
		Text:        tt.Text,
		Done:        tt.Done,
		Description: pgxconv.Text(tt.Description),
		Start:       pgxconv.Timestamp(tt.Start),
		Finish:      pgxconv.Timestamp(tt.Finish),
	})
	if err != nil {
		return models.TimetableTask{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return dtoTimetableTask(addedTimetable), nil
}

func (r *Repository) ListTimetableTasks(ctx context.Context, userID int) ([]models.TimetableTask, error) {
	op := fmt.Sprintf("list timetable tasks userID{%v} %%w", userID)
	tt, err := r.q.ListTimetableTasks(ctx, int32(userID))
	if err != nil {
		return nil, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return dtoTimetableTasks(tt), nil
}

func (r *Repository) DeleteTimetableTask(ctx context.Context, timetableTaskID, userID int) error {
	op := fmt.Sprintf("delete timetable tasks timetableTaskID{%v} userID{%v} %%w", timetableTaskID, userID)
	amount, err := r.q.DeleteTimetableTask(ctx, queries.DeleteTimetableTaskParams{
		ID:     int32(timetableTaskID),
		UserID: int32(userID),
	})

	if err != nil {
		return fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}
	if amount == 0 {
		return fmt.Errorf(op, serverrors.NewNoDeletionsError("timetable task"))
	}

	return nil
}

func (r *Repository) ListTimetableTasksInPeriod(ctx context.Context, userID int, from, to time.Time) ([]models.TimetableTask, error) {
	op := fmt.Sprintf("list timetable tasks userID{%v} from{%v} to{%v}: %%w", userID, from, to)
	tts, err := r.q.GetTimetableTasksInPeriod(ctx, queries.GetTimetableTasksInPeriodParams{
		UserID:   int32(userID),
		FromTime: pgxconv.Timestamp(from),
		ToTime:   pgxconv.Timestamp(to),
	})
	if err != nil {
		return nil, fmt.Errorf(op, err)
	}

	return dtoTimetableTasks(tts), nil
}

func (r *Repository) GetTimetableTask(ctx context.Context, timetableTaskID, userID int) (models.TimetableTask, error) {
	op := fmt.Sprintf("get timetable tasks timetableTaskID{%v} userID{%v} %%w", timetableTaskID, userID)
	tt, err := r.q.GetTimetableTask(ctx, queries.GetTimetableTaskParams{
		ID:     int32(timetableTaskID),
		UserID: int32(userID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.TimetableTask{}, fmt.Errorf(op, serverrors.NewNotFoundError(err, "timetable task"))
		}

		return models.TimetableTask{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return dtoTimetableTask(tt), nil
}

func (r *Repository) UpdateTimetableTask(ctx context.Context, tt models.TimetableTask) (models.TimetableTask, error) {
	op := "update timetable task: %w"
	updatedTT, err := r.q.UpdateTimetableTask(ctx, queries.UpdateTimetableTaskParams{
		ID:          int32(tt.ID),
		UserID:      int32(tt.UserID),
		Text:        tt.Text,
		Description: pgxconv.Text(tt.Description),
		Start:       pgxconv.Timestamp(tt.Start),
		Finish:      pgxconv.Timestamp(tt.Finish),
		Done:        tt.Done,
	})
	if err != nil {
		return models.TimetableTask{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return dtoTimetableTask(updatedTT), nil
}
