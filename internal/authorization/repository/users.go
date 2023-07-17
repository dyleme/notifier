package repository

import (
	"context"

	"github.com/Dyleme/Notifier/internal/authorization/repository/queries"
	"github.com/Dyleme/Notifier/internal/authorization/service"
)

func (r *Repository) CreateUser(ctx context.Context, input service.CreateUserInput) (int, error) {
	id, err := r.q.AddUser(ctx, queries.AddUserParams{
		Email:        input.Email,
		Nickname:     input.NickName,
		PasswordHash: input.Password,
	})
	if err != nil {
		return 0, err
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
