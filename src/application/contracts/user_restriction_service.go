package contracts

import "goproxy/domain/aggregates"

type UserRestrictionService interface {
	IsRestricted(user aggregates.User) bool
	AddToRestrictionList(user aggregates.User) error
	RemoveFromRestrictionList(user aggregates.User) error
}
