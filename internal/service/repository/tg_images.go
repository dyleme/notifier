package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	trmpgx "github.com/avito-tech/go-transaction-manager/pgxv5"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Dyleme/Notifier/internal/domain"
	"github.com/Dyleme/Notifier/internal/domain/apperr"
	"github.com/Dyleme/Notifier/internal/service/repository/queries/goqueries"
)

type TgImagesRepository struct {
	q      *goqueries.Queries
	cache  Cache
	getter *trmpgx.CtxGetter
	db     *pgxpool.Pool
}

type Cache interface {
	Get(key string, obj any) error
	Delete(key string) error
	Add(key string, obj any) error
}

func NewTGImagesRepository(db *pgxpool.Pool, getter *trmpgx.CtxGetter, cache Cache) *TgImagesRepository {
	return &TgImagesRepository{
		q:      goqueries.New(),
		cache:  cache,
		getter: getter,
		db:     db,
	}
}

func newTgImageKey(filename string) string {
	return "tg_images_filename=" + filename
}

func (t TgImagesRepository) Add(ctx context.Context, filename, tgFileID string) error {
	op := "TgImagesRepository.Add: %w"

	tx := t.getter.DefaultTrOrDB(ctx, t.db)
	tgImage, err := t.q.AddTgImage(ctx, tx, goqueries.AddTgImageParams{
		Filename: filename,
		TgFileID: tgFileID,
	})
	if err != nil {
		if intersection, isUnique := uniqueError(err, []string{"filename"}); isUnique {
			return fmt.Errorf(op, apperr.NewUniqueError(intersection, filename))
		}

		return fmt.Errorf(op, apperr.NewRepositoryError(err))
	}

	if err = t.cache.Add(newTgImageKey(filename), tgImage); err != nil {
		return fmt.Errorf(op, apperr.NewRepositoryError(err))
	}

	return nil
}

func uniqueError(err error, columnNames []string) (string, bool) {
	var pgerr *pgconn.PgError
	if errors.As(err, &pgerr) {
		if pgerr.Code == pgerrcode.UniqueViolation {
			for _, columnName := range columnNames {
				if strings.Contains(pgerr.Detail, columnName) {
					return columnName, true
				}
			}

			return "", true
		}
	}

	return "", false
}

func (t TgImagesRepository) Get(ctx context.Context, filename string) (domain.TgImage, error) {
	op := "TgImagesRepository.Get: %w"

	var tgImage goqueries.TgImage
	if err := t.cache.Get(newTgImageKey(filename), &tgImage); err == nil { // err == nil
		return dtoTgImage(tgImage), nil
	}

	tx := t.getter.DefaultTrOrDB(ctx, t.db)
	tgImage, err := t.q.GetTgImage(ctx, tx, filename)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.TgImage{}, fmt.Errorf(op, apperr.NewNotFoundError(err, "tg image"))
		}

		return domain.TgImage{}, fmt.Errorf(op, apperr.NewRepositoryError(err))
	}

	return domain.TgImage{
		ID:       int(tgImage.ID),
		Filename: tgImage.Filename,
		TgFileID: tgImage.TgFileID,
	}, nil
}

func dtoTgImage(tgImage goqueries.TgImage) domain.TgImage {
	return domain.TgImage{
		ID:       int(tgImage.ID),
		Filename: tgImage.Filename,
		TgFileID: tgImage.TgFileID,
	}
}
