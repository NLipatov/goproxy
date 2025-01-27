package application

import (
	"goproxy/domain/aggregates"
	"goproxy/domain/dataobjects"
	"goproxy/domain/events"
)

type Repository[T any] interface {
	GetById(id int) (T, error)
	Create(entity T) (int, error)
	Update(entity T) error
	Delete(entity T) error
}

type UserRepository interface {
	Repository[aggregates.User]
	GetByUsername(username string) (aggregates.User, error)
	GetByEmail(email string) (aggregates.User, error)
}

type EventRepository interface {
	GetById(id int) (events.OutboxEvent, error)
	Create(event events.OutboxEvent) (int, error)
	Update(event events.OutboxEvent) error
	Delete(event events.OutboxEvent) error
}

type PlanRepository interface {
	Repository[aggregates.Plan]
	GetAll() ([]aggregates.Plan, error)
	GetAllWithFeatures() ([]aggregates.Plan, error)
	GetByName(name string) (aggregates.Plan, error)
	GetByNameWithFeatures(name string) (aggregates.Plan, error)
	GetByIdWithFeatures(id int) (aggregates.Plan, error)
}

type PlanPriceRepository interface {
	Repository[dataobjects.PlanPrice]
	GetAllWithPlanId(planId int) ([]dataobjects.PlanPrice, error)
}

type UserPlanRepository interface {
	Repository[aggregates.UserPlan]
	GetUserActivePlan(userId int) (aggregates.UserPlan, error)
}

type PlanOfferRepository interface {
	GetOffers(planId int) ([]dataobjects.PlanLavatopOffer, error)
	Create(plo dataobjects.PlanLavatopOffer) (int, error)
	Update(plo dataobjects.PlanLavatopOffer) error
	Delete(plo dataobjects.PlanLavatopOffer) error
}
