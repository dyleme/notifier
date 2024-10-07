package userinfo

import (
	"context"
	"fmt"
	"time"

	"github.com/Dyleme/timecache"

	"github.com/Dyleme/Notifier/internal/authorization/service"
	"github.com/Dyleme/Notifier/internal/domains"
)

type UserRepoCache struct {
	userRepo UserRepo
	cache    *timecache.Cache[int, User]
}

type UserRepo interface {
	GetTGUserInfo(ctx context.Context, tgID int) (domains.User, error)
	CreateUser(ctx context.Context, input service.CreateUserInput) (domains.User, error)
	UpdateUserTime(ctx context.Context, id int, timezone domains.TimeZoneOffset, isDst bool) error
}

func NewUserRepoCache(userRepo UserRepo) *UserRepoCache {
	return &UserRepoCache{
		userRepo: userRepo,
		cache: timecache.NewWithConfig[int, User](timecache.Config{
			JanitorConfig: timecache.JanitorConfig{
				CleanPeriod:      time.Hour,
				StopJanitorEvery: 0,
			},
			StoreTime: time.Hour,
		}),
	}
}

type User struct {
	TGID  int
	ID    int
	Zone  int
	IsDST bool
}

func (u User) Location() *time.Location {
	return time.FixedZone("Temporary", u.Zone*int(time.Hour/time.Second))
}

func (u *UserRepoCache) GetUserInfo(ctx context.Context, tgID int) (User, error) {
	userID, err := u.cache.Get(tgID)
	if err == nil { // err equal nil
		return userID, nil
	}

	user, err := u.userRepo.GetTGUserInfo(ctx, tgID)
	if err != nil {
		return User{}, fmt.Errorf("get tg user info [tgID: %v]: %w", tgID, err)
	}

	u.cache.StoreDefDur(tgID, User{
		TGID:  user.TGID,
		ID:    user.ID,
		Zone:  user.TimeZoneOffset,
		IsDST: user.IsTimeZoneDST,
	})

	return userID, nil
}

func (u *UserRepoCache) UpdateUserTime(ctx context.Context, tgID, tzOffset int, isDST bool) error {
	op := "UserRepoCache.UpdateUserTime: %w"
	user, err := u.GetUserInfo(ctx, tgID)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	err = u.userRepo.UpdateUserTime(ctx, user.ID, domains.TimeZoneOffset(tzOffset), isDST)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	u.cache.Delete(tgID)

	return nil
}
