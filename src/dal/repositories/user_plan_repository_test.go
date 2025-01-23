package repositories

import (
	"database/sql"
	_ "github.com/lib/pq"
	"goproxy/dal/cache_serialization"
	"goproxy/dal/repositories/mocks"
	"goproxy/domain/aggregates"
	"os"
	"testing"
	"time"
)

func TestUserPlanRepository(t *testing.T) {
	setEnvErr := os.Setenv("DB_DATABASE", "plans")
	if setEnvErr != nil {
		t.Fatal(setEnvErr)
	}

	defer func() {
		_ = os.Unsetenv("DB_DATABASE")
	}()

	db, cleanup := prepareDb(t)
	defer cleanup()
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	planRepoCache := mocks.NewMockCacheWithTTL[[]cache_serialization.PlanDto]()
	planRepo := NewPlansRepository(db, planRepoCache)
	userPlansRepo := NewUserPlanRepository(db)

	t.Run("GetById", func(t *testing.T) {
		planId := insertTestPlan(planRepo, t)
		userPlanId := insertTestUserPlan(userPlansRepo, planId, t)
		userPlan, err := userPlansRepo.GetById(userPlanId)
		assertNoError(t, err, "Failed to load user plan by Id")
		loadedUserPlan, err := userPlansRepo.GetById(userPlan.Id())
		assertNoError(t, err, "Failed to load user plan by Id")
		assertUserPlansEqual(t, userPlan, loadedUserPlan)
	})

	t.Run("Create", func(t *testing.T) {
		planId := insertTestPlan(planRepo, t)
		userPlanId := insertTestUserPlan(userPlansRepo, planId, t)
		userPlan, err := userPlansRepo.GetById(userPlanId)
		assertNoError(t, err, "Failed to load user plan by Id")
		loadedUserPlan, err := userPlansRepo.GetById(userPlan.Id())
		assertNoError(t, err, "Failed to load inserted user plan")
		assertUserPlansEqual(t, userPlan, loadedUserPlan)
	})

	t.Run("Update", func(t *testing.T) {
		planId := insertTestPlan(planRepo, t)
		userPlanId := insertTestUserPlan(userPlansRepo, planId, t)
		userPlan, err := userPlansRepo.GetById(userPlanId)
		assertNoError(t, err, "Failed to load user plan by Id")

		newPlanId := insertTestPlan(planRepo, t)
		newUserId := userPlan.UserId() + 1
		newValidTo := userPlan.ValidTo().Add(time.Second * 5)
		updatedUserPlan, _ := aggregates.NewUserPlan(userPlanId, newUserId, newPlanId, newValidTo, time.Now())
		assertNoError(t, userPlansRepo.Update(updatedUserPlan), "Failed to update user plan")

		loadedUserPlan, err := userPlansRepo.GetById(userPlan.Id())
		assertNoError(t, err, "Failed to load updated user plan")
		assertUserPlansNotEqual(t, userPlan, loadedUserPlan)
	})

	t.Run("Delete", func(t *testing.T) {
		planId := insertTestPlan(planRepo, t)
		userPlanId := insertTestUserPlan(userPlansRepo, planId, t)
		userPlan, err := userPlansRepo.GetById(userPlanId)
		assertNoError(t, err, "Failed to load user plan by Id")

		assertNoError(t, userPlansRepo.Delete(userPlan), "Failed to delete user plan")

		_, err = userPlansRepo.GetById(userPlan.Id())
		if err == nil {
			t.Error("Expected error when loading deleted user plan")
		}
	})
}

func insertTestUserPlan(repo *UserPlanRepository, planId int, t *testing.T) int {
	userId := 1
	validTo := time.Now().Add(30 * 24 * time.Hour)
	createdAt := time.Now()

	userPlan, err := aggregates.NewUserPlan(-1, userId, planId, validTo, createdAt)
	assertNoError(t, err, "Failed to create test user plan lavatopaggregates")

	id, err := repo.Create(userPlan)
	assertNoError(t, err, "Failed to insert test user plan")
	return id
}

func assertUserPlansEqual(t *testing.T, expected, actual aggregates.UserPlan) {
	if expected.UserId() != actual.UserId() {
		t.Errorf("Expected UserId %d, got %d", expected.UserId(), actual.UserId())
	}
	if expected.PlanId() != actual.PlanId() {
		t.Errorf("Expected PlanId %d, got %d", expected.PlanId(), actual.PlanId())
	}
	if !expected.ValidTo().Equal(actual.ValidTo()) {
		t.Errorf("Expected ValidTo %s, got %s", expected.ValidTo(), actual.ValidTo())
	}
}

func assertUserPlansNotEqual(t *testing.T, expected, actual aggregates.UserPlan) {
	if expected.UserId() == actual.UserId() {
		t.Errorf("Unexpected equal UserId: %d", expected.UserId())
	}
	if expected.PlanId() == actual.PlanId() {
		t.Errorf("Unexpected equal PlanId: %d", expected.PlanId())
	}
	if expected.ValidTo().Equal(actual.ValidTo()) {
		t.Errorf("Unexpected equal ValidTo: %s", expected.ValidTo())
	}
}
