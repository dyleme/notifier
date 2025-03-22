package repository

import (
	"context"
	"errors"
	"fmt"
	"slices"

	trmpgx "github.com/avito-tech/go-transaction-manager/pgxv5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Dyleme/Notifier/internal/domain"
	"github.com/Dyleme/Notifier/internal/service/repository/queries/goqueries"
	"github.com/Dyleme/Notifier/internal/service/service"
	"github.com/Dyleme/Notifier/pkg/serverrors"
	"github.com/Dyleme/Notifier/pkg/utils"
)

type TagsRepository struct {
	q      *goqueries.Queries
	db     *pgxpool.Pool
	getter *trmpgx.CtxGetter
}

func NewTagsRepository(db *pgxpool.Pool, getter *trmpgx.CtxGetter) *TagsRepository {
	return &TagsRepository{
		q:      goqueries.New(),
		db:     db,
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
	tx := tr.getter.DefaultTrOrDB(ctx, tr.db)
	tags, err := tr.q.ListTags(ctx, tx, goqueries.ListTagsParams{
		UserID: int32(userID),
		Off:    int32(listParams.Offset),
		Lim:    int32(listParams.Limit),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []domain.Tag{}, nil
		}

		return nil, fmt.Errorf("list tags: %w", serverrors.NewRepositoryError(err))
	}

	return utils.DtoSlice(tags, dtoTag), nil
}

func (tr *TagsRepository) Get(ctx context.Context, tagID int) (domain.Tag, error) {
	tx := tr.getter.DefaultTrOrDB(ctx, tr.db)
	tag, err := tr.q.GetTag(ctx, tx, int32(tagID))
	if err != nil {
		return domain.Tag{}, fmt.Errorf("get tag[tagID=%v]: %w", tagID, serverrors.NewRepositoryError(err))
	}

	return dtoTag(tag), nil
}

func (tr *TagsRepository) Add(ctx context.Context, name string, userID int) (domain.Tag, error) {
	tx := tr.getter.DefaultTrOrDB(ctx, tr.db)
	createdTag, err := tr.q.AddTag(ctx, tx, goqueries.AddTagParams{
		Name:   name,
		UserID: int32(userID),
	})
	if err != nil {
		return domain.Tag{}, fmt.Errorf("add tag: %w", serverrors.NewRepositoryError(err))
	}

	return dtoTag(createdTag), nil
}

func (tr *TagsRepository) Delete(ctx context.Context, tagID int) error {
	tx := tr.getter.DefaultTrOrDB(ctx, tr.db)
	err := tr.q.DeleteTag(ctx, tx, int32(tagID))
	if err != nil {
		return fmt.Errorf("delete tag[tagID=%v]: %w", tagID, serverrors.NewRepositoryError(err))
	}

	return nil
}

func (tr *TagsRepository) Update(ctx context.Context, tagID int, name string) error {
	tx := tr.getter.DefaultTrOrDB(ctx, tr.db)

	err := tr.q.UpdateTag(ctx, tx, goqueries.UpdateTagParams{
		Name: name,
		ID:   int32(tagID),
	})
	if err != nil {
		return fmt.Errorf("update tag: %w", serverrors.NewRepositoryError(err))
	}

	return nil
}

func syncTags(ctx context.Context, tx trmpgx.Tr, q *goqueries.Queries, smthID, userID int, tags []domain.Tag) error {
	dbTags, err := q.ListTagsForSmth(ctx, tx, int32(smthID))
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("list tags for smth: %w", serverrors.NewRepositoryError(err))
		}
	}
	var tagIDsToDelete []int
	var tagsToInsert []domain.Tag

	dbTagIDs := utils.DtoSlice(dbTags, func(t goqueries.Tag) int { return int(t.ID) })
	tagIDs := utils.DtoSlice(tags, func(t domain.Tag) int { return t.ID })
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

	_, err = q.AddTagsToSmth(ctx, tx, utils.DtoSlice(tagsToInsert, func(t domain.Tag) goqueries.AddTagsToSmthParams {
		return goqueries.AddTagsToSmthParams{
			SmthID: int32(smthID),
			TagID:  int32(t.ID),
			UserID: int32(userID),
		}
	}))
	if err != nil {
		return fmt.Errorf("add tags to smth: %w", serverrors.NewRepositoryError(err))
	}

	err = q.DeleteTagsFromSmth(ctx, tx, goqueries.DeleteTagsFromSmthParams{
		SmthID: int32(smthID),
		TagIds: utils.DtoSlice(tagIDsToDelete, func(i int) int32 { return int32(i) }),
	})
	if err != nil {
		return fmt.Errorf("delete tags from smth: %w", serverrors.NewRepositoryError(err))
	}

	return nil
}
