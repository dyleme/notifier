package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/Dyleme/Notifier/internal/domain"
	"github.com/Dyleme/Notifier/internal/domain/apperr"
	"github.com/Dyleme/Notifier/internal/service/repository/queries/goqueries"
	"github.com/Dyleme/Notifier/internal/service/service"
	"github.com/Dyleme/Notifier/pkg/database/sqlconv"
	"github.com/Dyleme/Notifier/pkg/database/txmanager"
	"github.com/Dyleme/Notifier/pkg/utils/slice"
)

type BasicTaskRepository struct {
	q      *goqueries.Queries
	getter *txmanager.Getter
}

func NewBasicTaskRepository(getter *txmanager.Getter) *BasicTaskRepository {
	return &BasicTaskRepository{
		q:      goqueries.New(),
		getter: getter,
	}
}

func (er *BasicTaskRepository) dtoWithTags(bt goqueries.SingleTask, dbTags []goqueries.Tag) domain.BasicTask {
	basicTask := domain.BasicTask{
		ID:                 int(bt.ID),
		UserID:             int(bt.UserID),
		Text:               bt.Text,
		Description:        bt.Description.String,
		Start:              bt.Start,
		NotificationParams: bt.NotificationParams,
		Tags:               slice.Dto(dbTags, dtoTag),
	}

	return basicTask
}

func (er *BasicTaskRepository) Add(ctx context.Context, bt domain.BasicTask) (domain.BasicTask, error) {
	op := "add timetable task: %w"

	tx := er.getter.GetTx(ctx)
	addedTask, err := er.q.AddBasicTask(ctx, tx, goqueries.AddBasicTaskParams{
		UserID:             int64(bt.UserID),
		Text:               bt.Text,
		Start:              bt.Start,
		Description:        sqlconv.NullableString(bt.Description),
		NotificationParams: bt.NotificationParams,
	})
	if err != nil {
		return domain.BasicTask{}, fmt.Errorf(op, err)
	}

	for _, t := range bt.Tags {
		err = er.q.AddTagsToSmth(ctx, tx, goqueries.AddTagsToSmthParams{
			SmthID: addedTask.ID,
			TagID:  int64(t.ID),
			UserID: int64(bt.UserID),
		})
	}
	if err != nil {
		return domain.BasicTask{}, fmt.Errorf("add tags to smth: %w", err)
	}

	return er.Get(ctx, int(addedTask.ID))
}

func (er *BasicTaskRepository) List(ctx context.Context, userID int, params service.ListFilterParams) ([]domain.BasicTask, error) {
	tx := er.getter.GetTx(ctx)
	dbBasicTasks, err := er.q.ListBasicTasks(ctx, tx, goqueries.ListBasicTasksParams{
		UserID: int64(userID),
		Off:    int64(params.ListParams.Offset),
		Lim:    int64(params.ListParams.Limit),
		TagIds: slice.Dto(params.TagIDs, func(tagID int) int32 { return int32(tagID) }),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("list basic tasks: %w", err)
	}

	tasksIDs := slice.Dto(dbBasicTasks, func(t goqueries.ListBasicTasksRow) int64 { return t.SingleTask.ID })

	rows, err := er.q.ListTagsForSmths(ctx, tx, tasksIDs)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("list tags for smths: %w", err)
		}
	}

	basicTasks := make([]domain.BasicTask, 0, len(dbBasicTasks))
	for _, bt := range dbBasicTasks {
		var tags []goqueries.Tag
		for _, row := range rows {
			if row.SmthID == bt.SingleTask.ID {
				tags = append(tags, row.Tag)
			}
		}
		basicTasks = append(basicTasks, er.dtoWithTags(bt.SingleTask, tags))
	}

	return basicTasks, nil
}

func (er *BasicTaskRepository) Delete(ctx context.Context, taskID int) error {
	tx := er.getter.GetTx(ctx)
	deletedTasks, err := er.q.DeleteBasicTask(ctx, tx, int64(taskID))
	if err != nil {
		return fmt.Errorf("delete basic task: %w", err)
	}
	if len(deletedTasks) == 0 {
		return apperr.ErrNotFound
	}

	err = er.q.DeleteAllTagsForSmth(ctx, tx, int64(taskID))
	if err != nil {
		return fmt.Errorf("delete all tags for smth: %w", err)
	}

	return nil
}

func (er *BasicTaskRepository) Get(ctx context.Context, taskID int) (domain.BasicTask, error) {
	tx := er.getter.GetTx(ctx)
	dbBasicTask, err := er.q.GetBasicTask(ctx, tx, int64(taskID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.BasicTask{}, fmt.Errorf("get basic task: %w", apperr.NotFoundError{Object: "basic task"})
		}

		return domain.BasicTask{}, fmt.Errorf("get basic task: %w", err)
	}

	tags, err := er.q.ListTagsForSmth(ctx, tx, int64(taskID))
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return domain.BasicTask{}, fmt.Errorf("list tags for smth: %w", err)
		}
	}

	return er.dtoWithTags(dbBasicTask, tags), nil
}

func (er *BasicTaskRepository) Update(ctx context.Context, task domain.BasicTask) error {
	op := "update timetable task: %w"
	tx := er.getter.GetTx(ctx)
	err := er.q.UpdateBasicTask(ctx, tx, goqueries.UpdateBasicTaskParams{
		ID:                 int64(task.ID),
		Start:              task.Start,
		Text:               task.Text,
		Description:        sqlconv.NullableString(task.Description),
		NotificationParams: task.NotificationParams,
		UserID:             int64(task.UserID),
	})
	if err != nil {
		return fmt.Errorf(op, err)
	}

	err = syncTags(ctx, tx, er.q, task.UserID, task.ID, task.Tags)
	if err != nil {
		return fmt.Errorf("sync tags: %w", err)
	}

	return nil
}
