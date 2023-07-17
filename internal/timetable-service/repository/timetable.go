package repository

import (
	"context"
	"fmt"
	"time"

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
		return models.TimetableTask{}, err
	}

	return dtoTimetableTask(addedTimetable), nil
}

func (r *Repository) ListTimetableTasks(ctx context.Context, userID int) ([]models.TimetableTask, error) {
	tt, err := r.q.ListTimetableTasks(ctx, int32(userID))
	if err != nil {
		return nil, err
	}

	return dtoTimetableTasks(tt), nil
}

func (r *Repository) DeleteTimetableTask(ctx context.Context, timetableTaskID, userID int) error {
	amount, err := r.q.DeleteTimetableTask(ctx, queries.DeleteTimetableTaskParams{
		ID:     int32(timetableTaskID),
		UserID: int32(userID),
	})

	if err != nil {
		return err
	}
	if amount == 0 {
		return fmt.Errorf("no rows deleted")
	}

	return nil
}

func (r *Repository) ListTimetableTasksInPeriod(ctx context.Context, userID int, from, to time.Time) ([]models.TimetableTask, error) {
	tts, err := r.q.GetTimetableTasksInPeriod(ctx, queries.GetTimetableTasksInPeriodParams{
		UserID:   int32(userID),
		FromTime: pgxconv.Timestamp(from),
		ToTime:   pgxconv.Timestamp(to),
	})
	if err != nil {
		return nil, err
	}

	return dtoTimetableTasks(tts), nil
}

func (r *Repository) GetTimetableTask(ctx context.Context, timetableTaskID, userID int) (models.TimetableTask, error) {
	tt, err := r.q.GetTimetableTask(ctx, queries.GetTimetableTaskParams{
		ID:     int32(timetableTaskID),
		UserID: int32(userID),
	})
	if err != nil {
		return models.TimetableTask{}, err
	}

	return dtoTimetableTask(tt), nil
}

// func (r *Repository) UpdateTimetableTaskWithFunc(ctx context.Context, id int, updateFunc func(tt models.TimetableTask) models.TimetableTask) error {
// 	err := r.inTx(ctx, defaultTxOpts, func(q *queries.Queries) error {
// 		tt, err := q.GetTimetableTasks(ctx, int32(id))
// 		if err != nil {
// 			return err
// 		}
//
// 		updatedTimetableTask := updateFunc(dtoTimetableTask(tt))
//
// 		err = q.UpdateTimetableTask(ctx, queries.UpdateTimetableTaskParams{
// 			ID:      int32(updatedTimetableTask.ID),
// 			UserID:  int32(updatedTimetableTask.UserID),
// 			Start:   updatedTimetableTask.Start,
// 			Finish:  updatedTimetableTask.Finish,
// 			Text: updatedTimetableTask.Text,
// 			Done:    updatedTimetableTask.Done,
// 		})
// 		if err != nil {
// 			return err
// 		}
//
// 		return nil
// 	})
//
// 	if err != nil {
// 		return err
// 	}
//
// 	return nil
// }

func (r *Repository) UpdateTimetableTask(ctx context.Context, tt models.TimetableTask) (models.TimetableTask, error) {
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
		return models.TimetableTask{}, err
	}

	return dtoTimetableTask(updatedTT), nil
}
