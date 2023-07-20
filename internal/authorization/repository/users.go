package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Dyleme/Notifier/internal/authorization/repository/queries"
	"github.com/Dyleme/Notifier/internal/authorization/service"
	"github.com/Dyleme/Notifier/internal/lib/serverrors"
)

func (r *Repository) CreateUser(ctx context.Context, input service.CreateUserInput) (int, error) {
	op := "create user: %w"
	id, err := r.q.AddUser(ctx, queries.AddUserParams{
		Email:        input.Email,
		Nickname:     input.NickName,
		PasswordHash: input.Password,
	})
	if err != nil {
		if pgerr, ok := err.(*pgconn.PgError); ok {
			if pgerr.Code == pgerrcode.UniqueViolation {
				if strings.Contains(pgerr.Detail, "nickname") {
					return 0, fmt.Errorf(op, serverrors.NewUniqueError("nickname", input.NickName))
				}
				if strings.Contains(pgerr.Detail, "email") {
					return 0, fmt.Errorf(op, serverrors.NewUniqueError("email", input.Email))
				}
			}
		}

		return 0, fmt.Errorf(op, serverrors.NewRepositoryError(err))
	}

	return int(id), nil
}

func (r *Repository) GetPasswordHashAndID(ctx context.Context, authName string) (hash []byte, userID int, err error) {
	out, err := r.q.GetLoginParameters(ctx, authName)
	if err != nil {
		return nil, 0, err
	}

	return []byte(out.PasswordHash), int(out.ID), nil
}
