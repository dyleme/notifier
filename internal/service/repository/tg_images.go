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

	"github.com/Dyleme/Notifier/internal/domains"
	"github.com/Dyleme/Notifier/internal/service/repository/queries/goqueries"
	"github.com/Dyleme/Notifier/internal/service/service"
	"github.com/Dyleme/Notifier/pkg/serverrors"
)

type TgImagesRepository struct {
	q      *goqueries.Queries
	cache  Cache
	getter *trmpgx.CtxGetter
	db     *pgxpool.Pool
}

func (r *Repository) TgImages() service.TgImagesRepository {
	return r.tgImagesRepository
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
			return fmt.Errorf(op, serverrors.NewUniqueError(intersection, filename))
		}

		return fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	if err = t.cache.Add(newTgImageKey(filename), tgImage); err != nil {
		return fmt.Errorf(op, serverrors.NewRepositoryError(err))
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

func (t TgImagesRepository) Get(ctx context.Context, filename string) (domains.TgImage, error) {
	op := "TgImagesRepository.Get: %w"

	var tgImage goqueries.TgImage
	if err := t.cache.Get(newTgImageKey(filename), &tgImage); err == nil { // err == nil
		return dtoTgImage(tgImage), nil
	}

	tx := t.getter.DefaultTrOrDB(ctx, t.db)
	tgImage, err := t.q.GetTgImage(ctx, tx, filename)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domains.TgImage{}, fmt.Errorf(op, serverrors.NewNotFoundError(err, "tg image"))
		}

		return domains.TgImage{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return domains.TgImage{
		ID:       int(tgImage.ID),
		Filename: tgImage.Filename,
		TgFileID: tgImage.TgFileID,
	}, nil
}

func dtoTgImage(tgImage goqueries.TgImage) domains.TgImage {
	return domains.TgImage{
		ID:       int(tgImage.ID),
		Filename: tgImage.Filename,
		TgFileID: tgImage.TgFileID,
	}
}
