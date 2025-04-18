package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Dyleme/Notifier/internal/authorization/repository/queries/goqueries"
	"github.com/Dyleme/Notifier/internal/authorization/service"
	"github.com/Dyleme/Notifier/internal/domain"
	"github.com/Dyleme/Notifier/internal/domain/apperr"
	"github.com/Dyleme/Notifier/pkg/database/pgxconv"
	"github.com/Dyleme/Notifier/pkg/utils/slice"
)

func (r *Repository) Create(ctx context.Context, input service.CreateUserInput) (domain.User, error) {
	tx := r.getter.DefaultTrOrDB(ctx, r.db)
	op := "Repository.Create: %w"
	user, err := r.q.AddUser(ctx, tx, goqueries.AddUserParams{
		TgID:       int32(input.TGID),
		TgNickname: input.TGNickname,
	})
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
		TgNickname:     user.TgNickname,
		PasswordHash:   pgxconv.ByteSlice(user.PasswordHash),
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
	tx := r.getter.DefaultTrOrDB(ctx, r.db)
	out, err := r.q.FindUser(ctx, tx, goqueries.FindUserParams{
		TgNickname: nickname,
		TgID:       int32(tgID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, fmt.Errorf(op, apperr.ErrNotFound)
		}

		return domain.User{}, fmt.Errorf(op, err)
	}

	return domain.User{
		ID:             int(out.ID),
		TgNickname:     out.TgNickname,
		PasswordHash:   pgxconv.ByteSlice(out.PasswordHash),
		TGID:           int(out.TgID),
		TimeZoneOffset: int(out.TimezoneOffset),
		IsTimeZoneDST:  out.TimezoneDst,
	}, nil
}

func (r *Repository) Update(ctx context.Context, user domain.User) error {
	tx := r.getter.DefaultTrOrDB(ctx, r.db)
	err := r.q.UpdateUser(ctx, tx, goqueries.UpdateUserParams{
		TgNickname:     user.TgNickname,
		PasswordHash:   pgxconv.Text(string(user.PasswordHash)),
		TimezoneOffset: int32(user.TimeZoneOffset),
		TimezoneDst:    user.IsTimeZoneDST,
		TgID:           int32(user.TGID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperr.NotFoundError{Object: "user"}
		}

		return fmt.Errorf("find user: %w", err)
	}

	return nil
}

func (r *Repository) AddBindingAttempt(ctx context.Context, input service.BindingAttempt) error {
	tx := r.getter.DefaultTrOrDB(ctx, r.db)
	err := r.q.AddBindingAttempt(ctx, tx, goqueries.AddBindingAttemptParams{
		TgID:         int32(input.TGID),
		Code:         input.Code,
		PasswordHash: input.PasswordHash,
	})
	if err != nil {
		return fmt.Errorf("add binding attempt: %w", err)
	}

	return nil
}

func (r *Repository) GetLatestBindingAttempt(ctx context.Context, tgID int) (service.BindingAttempt, error) {
	tx := r.getter.DefaultTrOrDB(ctx, r.db)
	ba, err := r.q.GetLatestBindingAttempt(ctx, tx, int32(tgID))
	if err != nil {
		return service.BindingAttempt{}, fmt.Errorf("get latest binding attempt: %w", err)
	}

	return service.BindingAttempt{
		ID:           tgID,
		TGID:         int(ba.TgID),
		Code:         ba.Code,
		PasswordHash: ba.PasswordHash,
		Done:         ba.Done,
	}, nil
}

func (r *Repository) UpdateBindingAttemptStatus(ctx context.Context, baID int, done bool) error {
	tx := r.getter.DefaultTrOrDB(ctx, r.db)
	err := r.q.UpdateBindingAttempt(ctx, tx, goqueries.UpdateBindingAttemptParams{
		ID:   int32(baID),
		Done: done,
	})
	if err != nil {
		return fmt.Errorf("update binding attempt: %w", err)
	}

	return nil
}

func (r *Repository) GetNextTime(ctx context.Context) (time.Time, error) {
	tx := r.getter.DefaultTrOrDB(ctx, r.db)
	ts, err := r.q.GetNearestDailyNotificationTime(ctx, tx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Time{}, fmt.Errorf("get next time: %w", apperr.NotFoundError{Object: "time"})
		}

		return time.Time{}, fmt.Errorf("get next time: %w", err)
	}

	return pgxconv.OnlyTime(ts)
}

func (r *Repository) DailyNotificationsUsers(ctx context.Context, now time.Time) ([]domain.User, error) {
	tx := r.getter.DefaultTrOrDB(ctx, r.db)
	users, err := r.q.ListUsersToNotfiy(ctx, tx, pgxconv.PgOnlyTime(now))
	if err != nil {
		return nil, fmt.Errorf("get daily notifications users: %w", err)
	}

	return slice.Dto(users, func(u goqueries.User) domain.User {
		return domain.User{
			ID:             int(u.ID),
			TgNickname:     u.TgNickname,
			PasswordHash:   pgxconv.ByteSlice(u.PasswordHash),
			TGID:           int(u.TgID),
			TimeZoneOffset: int(u.TimezoneOffset),
			IsTimeZoneDST:  u.TimezoneDst,
		}
	}), nil
}
