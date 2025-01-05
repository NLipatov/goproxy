package infrastructure

import (
	"encoding/json"
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
		event, readErr := p.messageBus.Consume()
		if readErr == nil {
			p.consume(event)
		} else {
			log.Printf("Consumer error: %v (%v)\n", readErr, event)
		}
	}
}

func (p *PlanController) consume(outboxEvent *events.OutboxEvent) {
	var event events.UserConsumedTrafficEvent
	err := json.Unmarshal([]byte(outboxEvent.Payload), &event)
	if err != nil {
		log.Printf("Invalid event: %v", err)
		return
	}

	currentTraffic, err := p.cache.Get(p.cacheKey(event.UserId))
	if err != nil {
		//if cache miss
		if strings.Contains(err.Error(), "not found") {
			newUserTraffic, loadErr := p.loadFromDB(event.UserId)
			if loadErr != nil {
				return
			}
			currentTraffic = newUserTraffic
		} else {
			// produce user without plan event
			return
		}
	}

	currentTraffic.InBytes += event.InBytes
	currentTraffic.OutBytes += event.OutBytes
	currentTraffic.ActualizedAt = time.Now().UTC()

	err = p.cache.Set(p.cacheKey(event.UserId), currentTraffic)
	if err != nil {
		log.Printf("failed to update traffic: %v", err)
		return
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
	userPlanRow, userPlanRowFetchErr := p.getUserActivePlan(userId)
	if userPlanRowFetchErr != nil {
		if userPlanRowFetchErr.Error() == "sql: no rows in result set" {
			//Produce user without plan event
			log.Printf("failed to fetch user plan row: %s", userPlanRowFetchErr)
			return dto.UserTraffic{}, fmt.Errorf("user does not have a plan")
		}
	}

	activePlan, activePlanFetchErr := p.planRepository.GetById(userPlanRow.PlanId())
	if activePlanFetchErr != nil {
		//Produce user without plan event
		log.Printf("failed to fetch active plan: %s", activePlanFetchErr)
		return dto.UserTraffic{}, fmt.Errorf("user does not have a plan")
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
