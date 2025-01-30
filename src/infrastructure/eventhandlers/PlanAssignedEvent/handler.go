package PlanAssignedEvent

import (
	"encoding/json"
	"fmt"
	"goproxy/application/contracts"
	"goproxy/domain/aggregates"
	"goproxy/domain/dataobjects"
	"goproxy/domain/events"
	"goproxy/infrastructure/dto"
	"io"
	"net/http"
	"time"
)

type PlanAssignedHandler struct {
	messageBus         contracts.MessageBusService
	userPlanCache      contracts.CacheWithTTL[dataobjects.UserPlan]
	userPlanRepository contracts.UserPlanRepository
	userRepository     contracts.UserRepository
	planRepository     contracts.PlanRepository
	trafficCache       contracts.CacheWithTTL[dataobjects.UserTraffic]
	userApiHost        string
}

func NewPlanAssignedHandler(messageBus contracts.MessageBusService,
	userPlanCache contracts.CacheWithTTL[dataobjects.UserPlan],
	userPlanRepository contracts.UserPlanRepository,
	userRepository contracts.UserRepository,
	planRepository contracts.PlanRepository,
	trafficCache contracts.CacheWithTTL[dataobjects.UserTraffic],
	userApiHost string) *PlanAssignedHandler {
	return &PlanAssignedHandler{
		messageBus,
		userPlanCache,
		userPlanRepository,
		userRepository,
		planRepository,
		trafficCache,
		userApiHost,
	}
}

func (p *PlanAssignedHandler) Handle(payload string) error {
	var event events.PlanAssigned
	err := json.Unmarshal([]byte(payload), &event)
	if err != nil {
		return fmt.Errorf("invalid event: %s", err)
	}

	resp, err := http.Get(fmt.Sprintf("%s/users/get?email=%s", p.userApiHost, event.UserEmail))
	if err != nil {
		return fmt.Errorf("failed to fetch user id: %v", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, bodyErr := io.ReadAll(resp.Body)
	if bodyErr != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	var userResult dto.GetUserResult
	deserializationErr := json.Unmarshal(body, &userResult)
	if deserializationErr != nil {
		return fmt.Errorf("failed to deserialize user result: %v", deserializationErr)
	}

	plan, planErr := p.planRepository.GetById(event.PlanId)
	if planErr != nil {
		return fmt.Errorf("could not get plan by id: %d", event.PlanId)
	}

	newPlan, newPlanErr := aggregates.NewUserPlan(-1, userResult.Id, plan.Id(),
		time.Now().UTC().Add(time.Hour*24*time.Duration(plan.DurationDays())), time.Now().UTC())
	if newPlanErr != nil {
		return fmt.Errorf("could not create new user plan: %v", newPlanErr)
	}

	_, createNewPlanErr := p.userPlanRepository.Create(newPlan)
	if createNewPlanErr != nil {
		return fmt.Errorf("could not create new plan record in user_plan table: %d", event.PlanId)
	}

	_ = p.userPlanCache.Expire(fmt.Sprintf("user:%d:plan", userResult.Id), time.Nanosecond)
	_ = p.trafficCache.Expire(fmt.Sprintf("user:%d:traffic", userResult.Id), time.Nanosecond)
	_ = p.trafficCache.Expire(fmt.Sprintf("user:%d:restricted", userResult.Id), time.Nanosecond)

	return nil
}
