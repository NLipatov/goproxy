package services

import (
	"fmt"
	"goproxy/application"
	"goproxy/domain/aggregates"
	"strings"
	"time"
)

type UserRestrictionService struct {
	cache application.CacheWithTTL[bool]
}

func NewUserRestrictionService() *UserRestrictionService {
	return &UserRestrictionService{}
}

func (u *UserRestrictionService) UseCache(cache application.CacheWithTTL[bool]) *UserRestrictionService {
	u.cache = cache
	return u
}

func (u *UserRestrictionService) Build() (*UserRestrictionService, error) {
	if u.cache == nil {
		return nil, fmt.Errorf("cache must not be nil")
	}

	return u, nil
}

func (u *UserRestrictionService) IsRestricted(user aggregates.User) bool {
	_, err := u.cache.Get(u.toKey(user))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false
		}

		return false
	}

	return true
}

func (u *UserRestrictionService) AddToRestrictionList(user aggregates.User) error {
	err := u.cache.Set(u.toKey(user), true)
	return err
}

func (u *UserRestrictionService) RemoveFromRestrictionList(user aggregates.User) error {
	err := u.cache.Expire(u.toKey(user), time.Nanosecond)
	return err
}

func (u *UserRestrictionService) toKey(user aggregates.User) string {
	return user.Username()
}
