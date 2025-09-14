package repository

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/jackc/pgx/v5"

	"github.com/Dyleme/Notifier/internal/domain"
	"github.com/Dyleme/Notifier/internal/service/repository/queries/goqueries"
	"github.com/Dyleme/Notifier/internal/service/service"
	"github.com/Dyleme/Notifier/pkg/database/txmanager"
	"github.com/Dyleme/Notifier/pkg/utils/slice"
)

type TagsRepository struct {
	q      *goqueries.Queries
	getter *txmanager.Getter
}

func NewTagsRepository(getter *txmanager.Getter) *TagsRepository {
	return &TagsRepository{
		q:      goqueries.New(),
		getter: getter,
	}
}

func dtoTag(t goqueries.Tag) domain.Tag {
	return domain.Tag{
		ID:     int(t.ID),
		Name:   t.Name,
		UserID: int(t.UserID),
	}
}

func (tr *TagsRepository) List(ctx context.Context, userID int, listParams service.ListParams) ([]domain.Tag, error) {
	tx := tr.getter.GetTx(ctx)
	tags, err := tr.q.ListTags(ctx, tx, goqueries.ListTagsParams{
		UserID: int64(userID),
		Off:    int64(listParams.Offset),
		Lim:    int64(listParams.Limit),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []domain.Tag{}, nil
		}

		return nil, fmt.Errorf("list tags: %w", err)
	}

	return slice.Dto(tags, dtoTag), nil
}

func (tr *TagsRepository) Get(ctx context.Context, tagID int) (domain.Tag, error) {
	tx := tr.getter.GetTx(ctx)
	tag, err := tr.q.GetTag(ctx, tx, int64(tagID))
	if err != nil {
		return domain.Tag{}, fmt.Errorf("get tag[tagID=%v]: %w", tagID, err)
	}

	return dtoTag(tag), nil
}

func (tr *TagsRepository) Add(ctx context.Context, name string, userID int) (domain.Tag, error) {
	tx := tr.getter.GetTx(ctx)
	createdTag, err := tr.q.AddTag(ctx, tx, goqueries.AddTagParams{
		Name:   name,
		UserID: int64(userID),
	})
	if err != nil {
		return domain.Tag{}, fmt.Errorf("add tag: %w", err)
	}

	return dtoTag(createdTag), nil
}

func (tr *TagsRepository) Delete(ctx context.Context, tagID int) error {
	tx := tr.getter.GetTx(ctx)
	err := tr.q.DeleteTag(ctx, tx, int64(tagID))
	if err != nil {
		return fmt.Errorf("delete tag[tagID=%v]: %w", tagID, err)
	}

	return nil
}

func (tr *TagsRepository) Update(ctx context.Context, tagID int, name string) error {
	tx := tr.getter.GetTx(ctx)

	err := tr.q.UpdateTag(ctx, tx, goqueries.UpdateTagParams{
		Name: name,
		ID:   int64(tagID),
	})
	if err != nil {
		return fmt.Errorf("update tag: %w", err)
	}

	return nil
}

func syncTags(ctx context.Context, tx txmanager.DBTX, q *goqueries.Queries, smthID, userID int, tags []domain.Tag) error {
	dbTags, err := q.ListTagsForSmth(ctx, tx, int64(smthID))
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("list tags for smth: %w", err)
		}
	}
	var tagIDsToDelete []int
	var tagsToInsert []domain.Tag

	dbTagIDs := slice.Dto(dbTags, func(t goqueries.Tag) int { return int(t.ID) })
	tagIDs := slice.Dto(tags, func(t domain.Tag) int { return t.ID })
	for _, dbTagID := range dbTagIDs {
		if !slices.Contains(tagIDs, dbTagID) {
			tagIDsToDelete = append(tagIDsToDelete, dbTagID)
		}
	}

	for i, tagID := range tagIDs {
		if !slices.Contains(dbTagIDs, tagID) {
			tagsToInsert = append(tagsToInsert, tags[i])
		}
	}

	for _, tag := range tagsToInsert {
		err = q.AddTagsToSmth(ctx, tx, goqueries.AddTagsToSmthParams{
			SmthID: int64(smthID),
			TagID:  int64(tag.ID),
			UserID: int64(userID),
		})
		if err != nil {
			return fmt.Errorf("add tags to smth: %w", err)
		}
	}

	if len(tagIDsToDelete) != 0 {
		err = q.DeleteTagsFromSmth(ctx, tx, goqueries.DeleteTagsFromSmthParams{
			SmthID: int32(smthID),
			TagIds: slice.Dto(tagIDsToDelete, func(i int) int32 { return int32(i) }),
		})
		if err != nil {
			return fmt.Errorf("delete tags from smth: %w", err)
		}
	}

	return nil
}
