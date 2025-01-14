package application

import (
	"goproxy/domain/aggregates"
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
	GetByName(name string) (aggregates.Plan, error)
}

type UserPlanRepository interface {
	Repository[aggregates.UserPlan]
	GetUserActivePlan(userId int) (aggregates.UserPlan, error)
}
