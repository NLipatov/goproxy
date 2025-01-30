package use_cases

import (
	"fmt"
	"goproxy/application/contracts"
	"goproxy/domain/dataobjects"
	"log"
)

type UserPlanInfoUseCases struct {
	planRepository     contracts.PlanRepository
	userPlanRepository contracts.UserPlanRepository
	userRepository     contracts.UserRepository
	planCache          contracts.CacheWithTTL[dataobjects.UserPlan]
	trafficCache       contracts.CacheWithTTL[dataobjects.UserTraffic]
}

func NewUserPlanInfoUseCases(planRepository contracts.PlanRepository, userPlanRepository contracts.UserPlanRepository,
	userRepository contracts.UserRepository, planCache contracts.CacheWithTTL[dataobjects.UserPlan], trafficCache contracts.CacheWithTTL[dataobjects.UserTraffic]) UserPlanInfoUseCases {
	return UserPlanInfoUseCases{
		planRepository:     planRepository,
		userPlanRepository: userPlanRepository,
		userRepository:     userRepository,
		planCache:          planCache,
		trafficCache:       trafficCache,
	}
}

func (u *UserPlanInfoUseCases) FetchUserPlan(userId int) (dataobjects.UserPlan, error) {
	cached, cachedErr := u.planCache.Get(u.cachePlanKey(userId))
	if cachedErr == nil {
		return cached, nil
	}

	userPlan, userPlanErr := u.userPlanRepository.GetUserActivePlan(userId)
	if userPlanErr != nil {
		return dataobjects.UserPlan{}, userPlanErr
	}

	plan, planErr := u.planRepository.GetById(userPlan.PlanId())
	if planErr != nil {
		return dataobjects.UserPlan{}, planErr
	}

	userPlanData := dataobjects.UserPlan{
		Name:      plan.Name(),
		Bandwidth: plan.LimitBytes(),
	}

	cacheSetErr := u.planCache.Set(u.cachePlanKey(userId), userPlanData)
	if cacheSetErr != nil {
		log.Printf("failed to cache user plan: %v", cacheSetErr)
	}

	return userPlanData, nil
}

func (u *UserPlanInfoUseCases) FetchTrafficUsage(userId int) (int64, error) {
	traffic, trafficErr := u.trafficCache.Get(u.cacheTrafficKey(userId))
	if trafficErr != nil {
		return 0, trafficErr
	}

	return traffic.InBytes + traffic.OutBytes, nil
}

func (u *UserPlanInfoUseCases) cachePlanKey(userId int) string {
	return fmt.Sprintf("user:%d:plan", userId)
}

func (u *UserPlanInfoUseCases) cacheTrafficKey(userId int) string {
	return fmt.Sprintf("user:%d:traffic", userId)
}
