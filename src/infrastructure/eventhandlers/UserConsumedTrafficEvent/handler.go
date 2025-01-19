package UserConsumedTrafficEvent

import (
	"encoding/json"
	"errors"
	"fmt"
	"goproxy/application"
	"goproxy/domain"
	"goproxy/domain/aggregates"
	"goproxy/domain/dataobjects"
	"goproxy/domain/events"
	"goproxy/infrastructure/infraerrs"
	"log"
	"strings"
	"time"
)

type Handler struct {
	cache              application.CacheWithTTL[dataobjects.UserTraffic]
	userPlanRepository application.UserPlanRepository
	planRepository     application.PlanRepository
	messageBus         application.MessageBusService
}

func NewUserConsumedTrafficEventHandler(cache application.CacheWithTTL[dataobjects.UserTraffic],
	userPlanRepository application.UserPlanRepository,
	planRepository application.PlanRepository,
	messageBus application.MessageBusService) application.EventHandler {
	return &Handler{
		cache:              cache,
		userPlanRepository: userPlanRepository,
		planRepository:     planRepository,
		messageBus:         messageBus,
	}
}

func (u *Handler) Handle(payload string) error {
	var event events.UserConsumedTrafficEvent
	err := json.Unmarshal([]byte(payload), &event)
	if err != nil {
		return fmt.Errorf("invalid event: %v", err)
	}

	currentTraffic, err := u.cache.Get(u.cacheKey(event.UserId))
	if err != nil {
		newUserTraffic := dataobjects.UserTraffic{}
		//if cache miss - try load it form DB
		if strings.Contains(err.Error(), "not found") {
			dbResult, loadErr := u.loadFromDB(event.UserId)
			if loadErr != nil {
				if errors.Is(loadErr, infraerrs.UserWithNoPlan) {
					_ = u.produceUserWithNoPlanEvent(event.UserId)
				}
			}
			newUserTraffic = dbResult
		}

		currentTraffic = newUserTraffic
	}

	currentTraffic.InBytes += event.InBytes
	currentTraffic.OutBytes += event.OutBytes
	currentTraffic.ActualizedAt = time.Now().UTC()

	err = u.cache.Set(u.cacheKey(event.UserId), currentTraffic)
	if err != nil {
		log.Printf("cache update err: %v", err)
	}

	// if PlanLimitBytes is 0 -> no limit applied
	if currentTraffic.PlanLimitBytes == 0 {
		return nil
	}

	if currentTraffic.OutBytes+currentTraffic.InBytes > currentTraffic.PlanLimitBytes {
		produceErr := u.produceUserExceededTrafficLimitEvent(event.UserId)
		if produceErr != nil {
			log.Printf("could not produce user exceeded traffic limit event: %s", produceErr)
		}
	}

	return nil
}

func (u *Handler) cacheKey(userId int) string {
	return fmt.Sprintf("user:%d:traffic", userId)
}
func (u *Handler) loadFromDB(userId int) (dataobjects.UserTraffic, error) {
	activePlan, err := u.loadUserPlan(userId)
	if err != nil {
		return dataobjects.UserTraffic{}, err
	}
	userTraffic := dataobjects.UserTraffic{
		InBytes:        0,
		OutBytes:       0,
		PlanLimitBytes: activePlan.LimitBytes(),
		ActualizedAt:   time.Now().UTC(),
	}
	cacheErr := u.cache.Set(u.cacheKey(userId), userTraffic)
	if cacheErr != nil {
		log.Printf("failed to set to cache: %s", cacheErr)
	}

	_ = u.cache.Expire(u.cacheKey(userId), 24*time.Hour*time.Duration(activePlan.DurationDays()))

	return userTraffic, nil
}

func (u *Handler) loadUserPlan(userId int) (aggregates.Plan, error) {
	userPlanRow, userPlanRowFetchErr := u.getUserActivePlan(userId)
	if userPlanRowFetchErr != nil {
		log.Printf("failed to get user active plan: %s", userPlanRowFetchErr)
		return aggregates.Plan{}, infraerrs.UserWithNoPlan
	}

	activePlan, activePlanFetchErr := u.planRepository.GetById(userPlanRow.PlanId())
	if activePlanFetchErr != nil {
		return aggregates.Plan{}, infraerrs.UserWithNoPlan
	}

	return activePlan, nil
}

func (u *Handler) getUserActivePlan(userId int) (aggregates.UserPlan, error) {
	plan, err := u.userPlanRepository.GetUserActivePlan(userId)
	if err != nil {
		return plan, err
	}

	return plan, err
}

func (u *Handler) produceUserWithNoPlanEvent(userId int) error {
	userConsumedTrafficWithoutPlan := events.NewUserConsumedTrafficWithoutPlan(userId)
	data, serializationErr := json.Marshal(userConsumedTrafficWithoutPlan)
	if serializationErr != nil {
		log.Fatalf("failed to serialize user consumed a traffic without a plan event: %s", serializationErr)
		return serializationErr
	}

	outboxEvent, outboxEventValidationErr := events.NewOutboxEvent(0, string(data), false, "UserConsumedTrafficWithoutPlan")
	if outboxEventValidationErr != nil {
		return outboxEventValidationErr
	}

	produceErr := u.messageBus.Produce(fmt.Sprintf("%s", domain.PROXY), outboxEvent)
	if produceErr != nil {
		return produceErr
	}

	return nil
}

func (u *Handler) produceUserExceededTrafficLimitEvent(userId int) error {
	userExceededTrafficLimit := events.NewUserExceededTrafficLimitEvent(userId)
	data, serializationErr := json.Marshal(userExceededTrafficLimit)
	if serializationErr != nil {
		log.Fatalf("failed to serialize user exceeded traffic limit event: %s", serializationErr)
		return serializationErr
	}

	outboxEvent, outboxEventValidationErr := events.NewOutboxEvent(0, string(data), false, "UserExceededTrafficLimitEvent")
	if outboxEventValidationErr != nil {
		return outboxEventValidationErr
	}

	produceErr := u.messageBus.Produce(fmt.Sprintf("%s", domain.PROXY), outboxEvent)
	if produceErr != nil {
		return produceErr
	}

	return nil
}
