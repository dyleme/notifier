package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Dyleme/Notifier/internal/domain"
	"github.com/Dyleme/Notifier/internal/repository/queries/goqueries"
	"github.com/Dyleme/Notifier/internal/service"
	"github.com/Dyleme/Notifier/pkg/database/sqlconv"
	"github.com/Dyleme/Notifier/pkg/database/txmanager"
	"github.com/Dyleme/Notifier/pkg/log"
	"github.com/Dyleme/Notifier/pkg/utils/slice"
)

type TasksRepository struct {
	q      *goqueries.Queries
	getter *txmanager.Getter
}

func NewTasksRepository(getter *txmanager.Getter) *TasksRepository {
	return &TasksRepository{
		q:      goqueries.New(),
		getter: getter,
	}
}

func (r *TasksRepository) Add(ctx context.Context, task domain.Task) (domain.Task, error) {
	eventCreationParams, err := json.Marshal(task.EventCreationParams)
	if err != nil {
		return domain.Task{}, fmt.Errorf("marshal event creation params: %w", err)
	}
	tx := r.getter.GetTx(ctx)
	params := goqueries.AddTaskParams{
		Text:                task.Text,
		Description:         task.Description,
		UserID:              int64(task.UserID),
		Type:                string(task.Type),
		Time:                sqlconv.OnlyTimeFromDuration(task.Start),
		EventCreationParams: eventCreationParams,
	}
	log.Ctx(ctx).Debug("adding task", "task", task, "params", params)
	dbTask, err := r.q.AddTask(ctx, tx, params)
	if err != nil {
		return domain.Task{}, err
	}

	return r.dto(dbTask)
}

func (r *TasksRepository) Get(ctx context.Context, id, userID int) (domain.Task, error) {
	tx := r.getter.GetTx(ctx)
	task, err := r.q.GetTask(ctx, tx, goqueries.GetTaskParams{
		ID:     int64(id),
		UserID: int64(userID),
	})
	if err != nil {
		return domain.Task{}, err
	}

	return r.dto(task)
}

func (r *TasksRepository) List(ctx context.Context, userID int, taskType domain.TaskType, params service.ListParams) ([]domain.Task, error) {
	tx := r.getter.GetTx(ctx)
	tasks, err := r.q.ListTasks(ctx, tx, goqueries.ListTasksParams{
		UserID: int64(userID),
		Limit:  int64(params.Limit),
		Offset: int64(params.Offset),
		Type:   string(taskType),
	})
	if err != nil {
		return nil, err
	}

	return slice.DtoError(tasks, r.dto)
}

func (r *TasksRepository) Update(ctx context.Context, task domain.Task) error {
	eventCreationParams, err := json.Marshal(task.EventCreationParams)
	if err != nil {
		return fmt.Errorf("marshal event creation params: %w", err)
	}
	tx := r.getter.GetTx(ctx)
	err = r.q.UpdateTask(ctx, tx, goqueries.UpdateTaskParams{
		ID:                  int64(task.ID),
		Text:                task.Text,
		Description:         task.Description,
		UserID:              int64(task.UserID),
		Time:                sqlconv.OnlyTimeFromDuration(task.Start),
		EventCreationParams: eventCreationParams,
	})

	return err
}

func (r *TasksRepository) Delete(ctx context.Context, taskID, userID int) error {
	tx := r.getter.GetTx(ctx)
	_, err := r.q.DeleteTask(ctx, tx, goqueries.DeleteTaskParams{
		ID:     int64(taskID),
		UserID: int64(userID),
	})

	return err
}

func (r *TasksRepository) dto(t goqueries.Task) (domain.Task, error) {
	startTime, err := time.Parse(time.TimeOnly, t.Start)
	if err != nil {
		return domain.Task{}, fmt.Errorf("parse time: %w", err)
	}

	var eventCreationParams map[domain.CreationParamKey]any
	err = json.Unmarshal([]byte(t.EventCreationParams), &eventCreationParams)
	if err != nil {
		return domain.Task{}, fmt.Errorf("unmarshal event creation params: %w", err)
	}

	return domain.Task{
		ID:                  int(t.ID),
		CreatedAt:           t.CreatedAt,
		Text:                t.Text,
		Description:         t.Description,
		UserID:              int(t.UserID),
		Type:                domain.TaskType(t.Type),
		Start:               startTime.Sub(time.Time{}),
		EventCreationParams: eventCreationParams,
	}, nil
}
