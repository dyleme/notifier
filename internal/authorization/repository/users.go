package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Dyleme/Notifier/internal/authorization/repository/queries/goqueries"
	"github.com/Dyleme/Notifier/internal/authorization/service"
	"github.com/Dyleme/Notifier/internal/domain"
	"github.com/Dyleme/Notifier/internal/domain/apperr"
	"github.com/Dyleme/Notifier/pkg/database/sqliteconv"
)

func (r *Repository) Create(ctx context.Context, input service.CreateUserInput) (domain.User, error) {
	tx := r.getter.GetTx(ctx)
	op := "Repository.Create: %w"
	user, err := r.q.AddUser(ctx, tx, int64(input.TGID))
	if err != nil {
		if intersection, isUnique := uniqueError(err); isUnique {
			return domain.User{}, apperr.UniqueError{
				Name:  intersection,
				Value: input.TGNickname,
			}
		}

		return domain.User{}, fmt.Errorf(op, err)
	}

	return domain.User{
		ID:             int(user.ID),
		TGID:           int(user.TgID),
		TimeZoneOffset: 0,
		IsTimeZoneDST:  false,
	}, nil
}

func uniqueError(err error) (string, bool) {
	var pgerr *pgconn.PgError
	if errors.As(err, &pgerr) {
		if pgerr.Code == pgerrcode.UniqueViolation {
			if strings.Contains(pgerr.Detail, "email") {
				return "email", true
			}

			return "", true
		}
	}

	return "", false
}

func (r *Repository) Find(ctx context.Context, nickname string, tgID int) (domain.User, error) {
	op := "Repository.Find: %w"
	tx := r.getter.GetTx(ctx)
	out, err := r.q.FindUser(ctx, tx, int64(tgID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, fmt.Errorf(op, apperr.ErrNotFound)
		}

		return domain.User{}, fmt.Errorf(op, err)
	}

	return domain.User{
		ID:             int(out.ID),
		TGID:           int(out.TgID),
		TimeZoneOffset: int(out.TimezoneOffset),
		IsTimeZoneDST:  sqliteconv.ToBool(out.TimezoneDst),
	}, nil
}

func (r *Repository) Update(ctx context.Context, user domain.User) error {
	tx := r.getter.GetTx(ctx)
	err := r.q.UpdateUser(ctx, tx, goqueries.UpdateUserParams{
		TimezoneOffset: int64(user.TimeZoneOffset),
		TimezoneDst:    sqliteconv.BoolToInt(user.IsTimeZoneDST),
		TgID:           int64(user.TGID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperr.NotFoundError{Object: "user"}
		}

		return fmt.Errorf("find user: %w", err)
	}

	return nil
}
