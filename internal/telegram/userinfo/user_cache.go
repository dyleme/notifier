package userinfo

import (
	"context"
	"fmt"
	"time"

	"github.com/Dyleme/timecache"

	"github.com/Dyleme/Notifier/internal/authorization/service"
	"github.com/Dyleme/Notifier/internal/domain"
)

type UserRepoCache struct {
	userRepo UserRepo
	cache    *timecache.Cache[int, User]
}

type UserRepo interface {
	GetTGUserInfo(ctx context.Context, tgID int) (domain.User, error)
	CreateUser(ctx context.Context, input service.CreateUserInput) (domain.User, error)
	UpdateUserTime(ctx context.Context, id int, timezone domain.TimeZoneOffset, isDst bool) error
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

func (u *UserRepoCache) AddUser(ctx context.Context, tgID int, nickname string) (User, error) {
	domainUser, err := u.userRepo.CreateUser(ctx, service.CreateUserInput{
		TGNickname: nickname,
		TGID:       tgID,
	})
	if err != nil {
		return User{}, fmt.Errorf("repo create: %w", err)
	}

	user := User{
		TGID:  domainUser.TGID,
		ID:    domainUser.ID,
		Zone:  domainUser.TimeZoneOffset,
		IsDST: domainUser.IsTimeZoneDST,
	}

	u.cache.StoreDefDur(tgID, user)

	return user, nil
}

func (u *UserRepoCache) UpdateUserTime(ctx context.Context, tgID, tzOffset int, isDST bool) error {
	op := "UserRepoCache.UpdateUserTime: %w"
	user, err := u.GetUserInfo(ctx, tgID)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	err = u.userRepo.UpdateUserTime(ctx, user.ID, domain.TimeZoneOffset(tzOffset), isDST)
	if err != nil {
		return fmt.Errorf(op, err)
	}

	u.cache.Delete(tgID)

	return nil
}
