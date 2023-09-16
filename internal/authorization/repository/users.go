package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Dyleme/Notifier/internal/authorization/models"
	"github.com/Dyleme/Notifier/internal/authorization/repository/queries"
	"github.com/Dyleme/Notifier/internal/authorization/service"
	"github.com/Dyleme/Notifier/internal/lib/serverrors"
	"github.com/Dyleme/Notifier/internal/lib/sql/pgxconv"
)

func (r *Repository) Create(ctx context.Context, input service.CreateUserInput) (models.User, error) {
	op := "Repository.Create: %w"
	user, err := r.q.AddUser(ctx, queries.AddUserParams{
		Email:        pgxconv.Text(input.Email),
		PasswordHash: pgxconv.Text(input.Password),
		TgID:         pgxconv.Int4(input.TGID),
	})
	if err != nil {
		if intersection, isUnique := uniqueError(err); isUnique {
			return models.User{}, fmt.Errorf(op, serverrors.NewUniqueError(intersection, input.Email))
		}

		return models.User{}, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return models.User{
		ID:           int(user.ID),
		Email:        pgxconv.String(user.Email),
		PasswordHash: pgxconv.ByteSlice(user.PasswordHash),
		TGID:         pgxconv.Int(user.TgID),
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

func (r *Repository) Get(ctx context.Context, email string, tgID *int) (models.User, error) {
	op := "Repository.Get: %w"
	out, err := r.q.FindUser(ctx, queries.FindUserParams{
		Email: pgxconv.Text(email),
		TgID:  pgxconv.Int4(tgID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, fmt.Errorf(op, serverrors.NewNotFoundError(err, "user"))
		}

		return models.User{}, fmt.Errorf(op, err)
	}

	return models.User{
		ID:           int(out.ID),
		Email:        pgxconv.String(out.Email),
		PasswordHash: pgxconv.ByteSlice(out.PasswordHash),
		TGID:         pgxconv.Int(out.TgID),
	}, nil
}
