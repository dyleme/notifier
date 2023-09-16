package handler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Dyleme/timecache"

	"github.com/Dyleme/Notifier/internal/authorization/service"
	"github.com/Dyleme/Notifier/internal/lib/serverrors"
)

type UserRepoCache struct {
	userRepo service.UserRepo
	cache    *timecache.Cache[int, int]
}

func NewUserRepoCache(userRepo service.UserRepo) *UserRepoCache {
	return &UserRepoCache{
		userRepo: userRepo,
		cache: timecache.NewWithConfig[int, int](timecache.Config{
			JanitorConfig: timecache.JanitorConfig{
				CleanPeriod:      time.Hour,
				StopJanitorEvery: 0,
			},
			StoreTime: time.Hour,
		}),
	}
}

func (u *UserRepoCache) GetID(ctx context.Context, tgID int) (int, error) {
	op := "UserRepoCache.GetID: %w"
	userID, err := u.cache.Get(tgID)
	if err == nil { // err equal nil
		return userID, nil
	}

	err = u.userRepo.Atomic(ctx, func(ctx context.Context, userRepo service.UserRepo) error {
		user, err := userRepo.Get(ctx, "", &tgID) //nolint:govet //just err in tx
		if err == nil {                           // err equal nil
			userID = user.ID

			return nil
		}

		var notFoundErr serverrors.NotFoundError
		if !errors.As(err, &notFoundErr) {
			return err //nolint:wrapcheck // wrapping later
		}

		user, err = userRepo.Create(ctx, service.CreateUserInput{
			Email:    "",
			Password: "",
			TGID:     &tgID,
		})
		if err != nil {
			return err //nolint:wrapcheck // wrapping later
		}
		userID = user.ID

		return nil
	})
	if err != nil {
		return 0, fmt.Errorf(op, err)
	}

	u.cache.StoreDefDur(tgID, userID)

	return userID, nil
}
