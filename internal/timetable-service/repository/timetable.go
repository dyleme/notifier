package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Dyleme/Notifier/internal/lib/serverrors"
	"github.com/Dyleme/Notifier/internal/lib/sql/pgxconv"
	"github.com/Dyleme/Notifier/internal/lib/utils/dto"
	"github.com/Dyleme/Notifier/internal/timetable-service/models"
	"github.com/Dyleme/Notifier/internal/timetable-service/repository/queries"
	"github.com/Dyleme/Notifier/internal/timetable-service/service"
)

type TimetableTaskRepository struct {
	q *queries.Queries
}

func (r *Repository) TimetableTasks() service.TimetableTaskRepository {
	return &TimetableTaskRepository{q: r.q}
}

func dtoTimetableTask(t queries.TimetableTask) (models.TimetableTask, error) {
	return models.TimetableTask{
		ID:           int(t.ID),
		UserID:       int(t.UserID),
		TaskID:       int(t.TaskID),
		Text:         t.Text,
		Description:  pgxconv.String(t.Description),
		Start:        pgxconv.Time(t.Start),
		Finish:       pgxconv.Time(t.Finish),
		Done:         t.Done,
		Notification: t.Notification,
	}, nil
}

func (tr *TimetableTaskRepository) Add(ctx context.Context, tt models.TimetableTask) (models.TimetableTask, error) {
	op := "add timetable task: %w"
	addedTimetable, err := tr.q.AddTimetableTask(ctx, queries.AddTimetableTaskParams{
		UserID:       int32(tt.UserID),
		TaskID:       int32(tt.TaskID),
		Text:         tt.Text,
		Done:         tt.Done,
		Description:  pgxconv.Text(tt.Description),
		Start:        pgxconv.Timestamp(tt.Start),
		Finish:       pgxconv.Timestamp(tt.Finish),
		Notification: models.Notification{Sended: false, Params: nil},
	})
	if err != nil {
		return models.TimetableTask{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return dtoTimetableTask(addedTimetable)
}

func (tr *TimetableTaskRepository) List(ctx context.Context, userID int) ([]models.TimetableTask, error) {
	op := fmt.Sprintf("list timetable tasks userID{%v} %%w", userID)
	tt, err := tr.q.ListTimetableTasks(ctx, int32(userID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return dto.ErrorSlice(tt, dtoTimetableTask)
}

func (tr *TimetableTaskRepository) Delete(ctx context.Context, timetableTaskID, userID int) error {
	op := fmt.Sprintf("delete timetable tasks timetableTaskID{%v} userID{%v} %%w", timetableTaskID, userID)
	amount, err := tr.q.DeleteTimetableTask(ctx, queries.DeleteTimetableTaskParams{
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

func (tr *TimetableTaskRepository) ListInPeriod(ctx context.Context, userID int, from, to time.Time) ([]models.TimetableTask, error) {
	op := fmt.Sprintf("list timetable tasks userID{%v} from{%v} to{%v}: %%w", userID, from, to)
	tts, err := tr.q.GetTimetableTasksInPeriod(ctx, queries.GetTimetableTasksInPeriodParams{
		UserID:   int32(userID),
		FromTime: pgxconv.Timestamp(from),
		ToTime:   pgxconv.Timestamp(to),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return dto.ErrorSlice(tts, dtoTimetableTask)
}

func (tr *TimetableTaskRepository) Get(ctx context.Context, timetableTaskID, userID int) (models.TimetableTask, error) {
	op := fmt.Sprintf("get timetable tasks timetableTaskID{%v} userID{%v} %%w", timetableTaskID, userID)
	tt, err := tr.q.GetTimetableTask(ctx, queries.GetTimetableTaskParams{
		ID:     int32(timetableTaskID),
		UserID: int32(userID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.TimetableTask{}, fmt.Errorf(op, serverrors.NewNotFoundError(err, "timetable task"))
		}
		return models.TimetableTask{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return dtoTimetableTask(tt)
}

func (tr *TimetableTaskRepository) Update(ctx context.Context, tt models.TimetableTask) (models.TimetableTask, error) {
	op := "update timetable task: %w"
	updatedTT, err := tr.q.UpdateTimetableTask(ctx, queries.UpdateTimetableTaskParams{
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

	return dtoTimetableTask(updatedTT)
}

func (tr *TimetableTaskRepository) GetNotNotified(ctx context.Context) ([]models.TimetableTask, error) {
	op := "TimetableTaskRepository.GetNotNotified: %w"
	tasks, err := tr.q.GetTimetableReadyTasks(ctx)
	if err != nil {
		return nil, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	notNotified, err := dto.ErrorSlice(tasks, dtoTimetableTask)
	if err != nil {
		return nil, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return notNotified, nil
}

func (tr *TimetableTaskRepository) MarkNotified(ctx context.Context, ids []int) error {
	op := "TimetableTaskRepository.MarkNotified: %w"
	err := tr.q.MarkNotificationSended(ctx, dto.Slice(ids, func(i int) int32 {
		return int32(i)
	}))
	if err != nil {
		return fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return nil
}

func (tr *TimetableTaskRepository) UpdateNotificationParams(ctx context.Context, timetableTaskID, userID int, params models.NotificationParams) (models.NotificationParams, error) {
	op := "TimetableTaskRepository.UpdateNotificationParams: %w"
	bts, err := json.Marshal(params)
	if err != nil {
		return models.NotificationParams{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	p, err := tr.q.UpdateNotificationParams(ctx, queries.UpdateNotificationParamsParams{
		Params: bts,
		ID:     int32(timetableTaskID),
		UserID: int32(userID),
	})

	if err != nil {
		return models.NotificationParams{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	if p.Params == nil {
		return models.NotificationParams{}, fmt.Errorf(op, serverrors.NewRepositoryError(fmt.Errorf("params are nil after update")))
	}
	return *p.Params, nil
}
