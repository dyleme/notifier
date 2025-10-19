package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/dyleme/Notifier/internal/domain"
	"github.com/dyleme/Notifier/internal/domain/apperr"
	"github.com/dyleme/Notifier/internal/repository/queries/goqueries"
	"github.com/dyleme/Notifier/pkg/database/sqlconv"
	"github.com/dyleme/Notifier/pkg/database/txmanager"
)

type UsersRepository struct {
	q      *goqueries.Queries
	getter *txmanager.Getter
}

func NewUserRepository(getter *txmanager.Getter) *UsersRepository {
	return &UsersRepository{
		q:      goqueries.New(),
		getter: getter,
	}
}

func (r *UsersRepository) dto(dbUser goqueries.User) domain.User {
	return domain.User{
		ID:                        int(dbUser.ID),
		TGID:                      int(dbUser.TgID),
		TimeZoneOffset:            int(dbUser.TimezoneOffset),
		IsTimeZoneDST:             sqlconv.ToBool(dbUser.TimezoneDst),
		DefaultNotificationPeriod: time.Duration(dbUser.NotificationRetryPeriodS) * time.Second,
	}
}

func (r *UsersRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	tx := r.getter.GetTx(ctx)
	dbUser, err := r.q.CreateUser(ctx, tx, goqueries.CreateUserParams{
		TgID:                     int64(user.TGID),
		TimezoneOffset:           int64(user.TimeZoneOffset),
		TimezoneDst:              sqlconv.BoolToInt(user.IsTimeZoneDST),
		NotificationRetryPeriodS: int64(user.DefaultNotificationPeriod.Seconds()),
	})
	if err != nil {
		return domain.User{}, err
	}

	return r.dto(dbUser), nil
}

func (r *UsersRepository) Get(ctx context.Context, id int) (domain.User, error) {
	tx := r.getter.GetTx(ctx)
	dbUser, err := r.q.GetUser(ctx, tx, int64(id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, apperr.ErrNotFound
		}

		return domain.User{}, fmt.Errorf("find user: %w", err)
	}

	return r.dto(dbUser), nil
}

func (r *UsersRepository) GetByTgID(ctx context.Context, tgID int) (domain.User, error) {
	op := "Repository.Find: %w"
	tx := r.getter.GetTx(ctx)
	dbUser, err := r.q.GetUserByTgID(ctx, tx, int64(tgID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, fmt.Errorf(op, apperr.ErrNotFound)
		}

		return domain.User{}, fmt.Errorf(op, err)
	}

	return r.dto(dbUser), nil
}

func (r *UsersRepository) Update(ctx context.Context, user domain.User) error {
	tx := r.getter.GetTx(ctx)
	err := r.q.UpdateUser(ctx, tx, goqueries.UpdateUserParams{
		ID:                       int64(user.ID),
		TimezoneOffset:           int64(user.TimeZoneOffset),
		TimezoneDst:              sqlconv.BoolToInt(user.IsTimeZoneDST),
		NotificationRetryPeriodS: int64(user.DefaultNotificationPeriod / time.Second),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperr.NotFoundError{Object: "user"}
		}

		return fmt.Errorf("find user: %w", err)
	}

	return nil
}
