package infrastructure

import (
	"encoding/json"
	"errors"
	"fmt"
	"goproxy/application"
	"goproxy/domain/aggregates"
	"goproxy/domain/events"
	"goproxy/infrastructure/dto"
	"log"
	"os"
	"strings"
	"time"
)

var userWithNoPlan = errors.New("user with no plan")

type PlanController struct {
	userPlanRepo   application.UserPlanRepository
	planRepository application.PlanRepository
	cache          application.CacheWithTTL[dto.UserTraffic]
	messageBus     application.MessageBusService
}

func NewTrafficCollector() *PlanController {
	return &PlanController{}
}

func (p *PlanController) UseCache(cache application.CacheWithTTL[dto.UserTraffic]) *PlanController {
	p.cache = cache
	return p
}

func (p *PlanController) UseMessageBus(busService application.MessageBusService) *PlanController {
	p.messageBus = busService
	return p
}

func (p *PlanController) UsePlanRepository(repo application.PlanRepository) *PlanController {
	p.planRepository = repo
	return p
}

func (p *PlanController) UseUserPlanRepository(repo application.UserPlanRepository) *PlanController {
	p.userPlanRepo = repo
	return p
}

func (p *PlanController) Build() (*PlanController, error) {
	if p.messageBus == nil {
		return nil, fmt.Errorf("message bus must be set")
	}

	if p.cache == nil {
		return nil, fmt.Errorf("cache must be set")
	}

	if p.planRepository == nil {
		return nil, fmt.Errorf("plan repository bus must be set")
	}

	if p.userPlanRepo == nil {
		return nil, fmt.Errorf("user plan repository must be set")
	}

	return p, nil
}

func (p *PlanController) ProcessEvents() {
	defer func(messageBus application.MessageBusService) {
		_ = messageBus.Close()
	}(p.messageBus)

	userTrafficTopic := os.Getenv("KAFKA_TOPIC")
	topics := []string{userTrafficTopic}
	err := p.messageBus.Subscribe(topics)
	if err != nil {
		log.Fatalf("Failed to subscribe to topics: %s", err)
	}

	log.Printf("Subscribed to topics: %s", strings.Join(topics, ", "))

	for {
		event, consumeErr := p.messageBus.Consume()
		if consumeErr != nil {
			log.Printf("failed to consume from message bus: %s", consumeErr)
		}

		if event.EventType.Value() == "NewUserConsumedTrafficEvent" {
			p.consumeUserConsumedTraffic(event)
		}
	}
}

func (p *PlanController) consumeUserConsumedTraffic(outboxEvent *events.OutboxEvent) {
	var event events.UserConsumedTrafficEvent
	err := json.Unmarshal([]byte(outboxEvent.Payload), &event)
	if err != nil {
		log.Printf("Invalid event: %v", err)
		return
	}

	currentTraffic, err := p.cache.Get(p.cacheKey(event.UserId))
	if err != nil {
		newUserTraffic := dto.UserTraffic{}
		//if cache miss - try load it form DB
		if strings.Contains(err.Error(), "not found") {
			dbResult, loadErr := p.loadFromDB(event.UserId)
			if loadErr != nil {
				if errors.Is(loadErr, userWithNoPlan) {
					_ = p.produceUserWithNoPlanEvent(event.UserId)
				}
			}
			newUserTraffic = dbResult
		}

		currentTraffic = newUserTraffic
	}

	currentTraffic.InBytes += event.InBytes
	currentTraffic.OutBytes += event.OutBytes
	currentTraffic.ActualizedAt = time.Now().UTC()

	err = p.cache.Set(p.cacheKey(event.UserId), currentTraffic)
	if err != nil {
		log.Printf("cache update err: %v", err)
	}

	if currentTraffic.OutBytes+currentTraffic.InBytes > currentTraffic.PlanLimitBytes {

	}
}

func (p *PlanController) cacheKey(userId int) string {
	return fmt.Sprintf("user:%d:traffic:%s", userId, time.Now().UTC().Format("02-01-2006"))
}

func (p *PlanController) getUserActivePlan(userId int) (aggregates.UserPlan, error) {
	plan, err := p.userPlanRepo.GetUserActivePlan(userId)
	if err != nil {
		return plan, err
	}

	return plan, err
}

func (p *PlanController) loadFromDB(userId int) (dto.UserTraffic, error) {
	activePlan, err := p.loadUserPlan(userId)
	if err != nil {
		return dto.UserTraffic{}, err
	}
	userTraffic := dto.UserTraffic{
		InBytes:        0,
		OutBytes:       0,
		PlanLimitBytes: activePlan.LimitBytes(),
		ActualizedAt:   time.Now().UTC(),
	}
	cacheErr := p.cache.Set(p.cacheKey(userId), userTraffic)
	if cacheErr != nil {
		log.Printf("failed to set to cache: %s", cacheErr)
	}

	_ = p.cache.Expire(p.cacheKey(userId), 24*time.Hour*time.Duration(activePlan.DurationDays()))

	return userTraffic, nil
}

func (p *PlanController) loadUserPlan(userId int) (aggregates.Plan, error) {
	userPlanRow, userPlanRowFetchErr := p.getUserActivePlan(userId)
	if userPlanRowFetchErr != nil {
		return aggregates.Plan{}, userWithNoPlan
	}

	activePlan, activePlanFetchErr := p.planRepository.GetById(userPlanRow.PlanId())
	if activePlanFetchErr != nil {
		return aggregates.Plan{}, userWithNoPlan
	}

	return activePlan, nil
}

func (p *PlanController) produceUserWithNoPlanEvent(userId int) error {
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

	produceErr := p.messageBus.Produce("user-plans", outboxEvent)
	if produceErr != nil {
		return produceErr
	}

	return nil
}

func (p *PlanController) produceUserExceededTrafficLimitEvent(userId int) error {
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

	produceErr := p.messageBus.Produce("user-plans", outboxEvent)
	if produceErr != nil {
		return produceErr
	}

	return nil
}
