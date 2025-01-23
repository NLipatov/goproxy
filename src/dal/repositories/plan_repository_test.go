package repositories

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"goproxy/application"
	"goproxy/dal/cache_serialization"
	"goproxy/dal/repositories/mocks"
	"goproxy/domain/aggregates"
	"goproxy/domain/valueobjects"
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestPlansRepository(t *testing.T) {
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

	planRepositoryCache := mocks.NewMockCacheWithTTL[[]cache_serialization.PlanDto]()
	planRepository := NewPlansRepository(db, planRepositoryCache)

	t.Run("GetAll", func(t *testing.T) {
		// plan count is in [0;10] range
		planCount := rand.Intn(9) + 1
		expectedPlanIds := make([]int, planCount)

		for i := 0; i < planCount; i++ {
			expectedPlanIds[i] = insertTestPlan(planRepository, t)
		}

		plans, plansErr := planRepository.GetAll()
		if plansErr != nil {
			t.Fatal(plansErr)
		}

		actualPlanIds := make(map[int]bool)
		for _, v := range plans {
			actualPlanIds[v.Id()] = true
		}

		for _, v := range expectedPlanIds {
			if !actualPlanIds[v] {
				t.Errorf("plan %d not found in expected plans", v)
			}
		}
	})

	t.Run("GetAllWithFeatures", func(t *testing.T) {
		// plan count is in [0;10] range
		planCount := rand.Intn(9) + 1
		expectedPlanIds := make(map[int]int, planCount)
		expectedPlanFeatureCount := make(map[int]int, planCount)

		for i := 0; i < planCount; i++ {
			featureCount := i + 1
			expectedPlanIds[i] = insertTestPlanWithFeatures(db, planRepository, t, featureCount)
			expectedPlanFeatureCount[expectedPlanIds[i]] = featureCount
		}

		plans, plansErr := planRepository.GetAllWithFeatures()
		if plansErr != nil {
			t.Fatal(plansErr)
		}

		actualPlanIds := make(map[int]aggregates.Plan)
		for _, v := range plans {
			actualPlanIds[v.Id()] = v
		}

		for _, id := range expectedPlanIds {
			actualPlan, exists := actualPlanIds[id]
			if !exists {
				t.Errorf("plan %d not found in expected plans", actualPlan.Id())
			}

			expectedFeatureCount := expectedPlanFeatureCount[id]
			if len(actualPlan.Features()) != expectedFeatureCount {
				t.Fatalf("Expected plan %d to have %d features, gor %d features", id, expectedFeatureCount, len(actualPlan.Features()))
			}
		}
	})

	t.Run("GetByName", func(t *testing.T) {
		planId := insertTestPlan(planRepository, t)
		plan, err := planRepository.GetById(planId)
		assertNoError(t, err, "Failed to load plan by Id")
		loadedPlan, err := planRepository.GetByName(plan.Name())
		assertNoError(t, err, "Failed to load plan by Name")
		assertPlansEqual(t, plan, loadedPlan)
	})

	t.Run("GetById", func(t *testing.T) {
		planId := insertTestPlan(planRepository, t)
		plan, err := planRepository.GetById(planId)
		assertNoError(t, err, "Failed to load plan by Id")
		loadedPlan, err := planRepository.GetById(plan.Id())
		assertNoError(t, err, "Failed to load plan by Id")
		assertPlansEqual(t, plan, loadedPlan)
	})

	t.Run("Create", func(t *testing.T) {
		planId := insertTestPlan(planRepository, t)
		plan, err := planRepository.GetById(planId)
		assertNoError(t, err, "Failed to load plan by Id")
		loadedPlan, err := planRepository.GetByName(plan.Name())
		assertNoError(t, err, "Failed to load inserted plan")
		assertPlansEqual(t, plan, loadedPlan)
	})

	t.Run("Update", func(t *testing.T) {
		planId := insertTestPlan(planRepository, t)
		plan, err := planRepository.GetById(planId)
		assertNoError(t, err, "Failed to load plan by Id")

		updatedPlan, _ := aggregates.NewPlan(planId, "Updated Plan", 2000000, 60,
			make([]valueobjects.PlanFeature, 0))
		assertNoError(t, planRepository.Update(updatedPlan), "Failed to update plan")

		loadedPlan, err := planRepository.GetById(plan.Id())
		assertNoError(t, err, "Failed to load updated plan")
		assertPlansNotEqual(t, plan, loadedPlan)
	})

	t.Run("Delete", func(t *testing.T) {
		planId := insertTestPlan(planRepository, t)
		plan, err := planRepository.GetById(planId)
		assertNoError(t, err, "Failed to load plan by Id")

		assertNoError(t, planRepository.Delete(plan), "Failed to delete plan")

		_, err = planRepository.GetById(plan.Id())
		if err == nil {
			t.Error("Expected error when loading deleted plan")
		}
	})

	t.Run("GetByIdWithFeatures", func(t *testing.T) {
		featureCount := rand.Intn(9) + 1
		planId := insertTestPlanWithFeatures(db, planRepository, t, featureCount)
		plan, err := planRepository.GetByIdWithFeatures(planId)
		if err != nil {
			t.Fatal(err)
		}

		if len(plan.Features()) != featureCount {
			t.Fatalf("Expected plan %d to have %d features, gor %d features", plan.Id(), featureCount, len(plan.Features()))
		}
	})

	t.Run("GetByNameWithFeatures", func(t *testing.T) {
		featureCount := rand.Intn(9) + 1
		planId := insertTestPlanWithFeatures(db, planRepository, t, featureCount)
		plan, planErr := planRepository.GetById(planId)
		plan, err := planRepository.GetByNameWithFeatures(plan.Name())

		if len(plan.Features()) != featureCount {
			t.Fatalf("Expected plan %d to have %d features, gor %d features", plan.Id(), featureCount, len(plan.Features()))
		}

		assertNoError(t, planErr, "Failed to load plan by Id")
		assertNoError(t, err, "Failed to load plan with features by Id")
		assertNoError(t, err, "Failed to load plan with features by Name")
	})

}

func insertTestPlan(repo application.PlanRepository, t *testing.T) int {
	name := fmt.Sprintf("Test Plan %d", time.Now().UTC().UnixNano())
	plan, err := aggregates.NewPlan(-1, name, 1000000, 30, make([]valueobjects.PlanFeature, 0))
	assertNoError(t, err, "Failed to create test plan plan")

	id, err := repo.Create(plan)
	assertNoError(t, err, "Failed to insert test plan")
	return id
}

func assertPlansEqual(t *testing.T, expected, actual aggregates.Plan) {
	if expected.Name() != actual.Name() {
		t.Errorf("Expected Name %s, got %s", expected.Name(), actual.Name())
	}
	if expected.LimitBytes() != actual.LimitBytes() {
		t.Errorf("Expected LimitBytes %d, got %d", expected.LimitBytes(), actual.LimitBytes())
	}
	if expected.DurationDays() != actual.DurationDays() {
		t.Errorf("Expected DurationDays %d, got %d", expected.DurationDays(), actual.DurationDays())
	}
	if len(expected.Features()) != len(actual.Features()) {
		t.Errorf("Expected %d features, got %d", len(expected.Features()), len(actual.Features()))
	} else {
		for i, feature := range expected.Features() {
			if feature.Feature() != actual.Features()[i].Feature() {
				t.Errorf("Expected Feature %s, got %s", feature.Feature(), actual.Features()[i].Feature())
			}
		}
	}
}

func assertPlansNotEqual(t *testing.T, expected, actual aggregates.Plan) {
	if expected.Name() == actual.Name() {
		t.Errorf("Unexpected equal Names: %s", expected.Name())
	}
	if expected.LimitBytes() == actual.LimitBytes() {
		t.Errorf("Unexpected equal LimitBytes: %d", expected.LimitBytes())
	}
	if expected.DurationDays() == actual.DurationDays() {
		t.Errorf("Unexpected equal DurationDays: %d", expected.DurationDays())
	}
}

func insertTestPlanWithFeatures(db *sql.DB, repo application.PlanRepository, t *testing.T, featureCount int) int {
	name := fmt.Sprintf("Test Plan With Features %d", time.Now().UTC().UnixNano())
	plan, err := aggregates.NewPlan(-1, name, time.Now().UnixMilli(), time.Now().Day(), make([]valueobjects.PlanFeature, 0))
	assertNoError(t, err, "Failed to create test plan")

	planId, err := repo.Create(plan)
	assertNoError(t, err, "Failed to insert test plan")

	addFeaturesToPlan(db, t, planId, featureCount)

	return planId
}

func addFeaturesToPlan(db *sql.DB, t *testing.T, planId int, featureCount int) {
	for i := 0; i < featureCount; i++ {
		featureName := fmt.Sprintf("Feature %d %d", i, rand.Int63())
		featureDescription := fmt.Sprintf("Feature %d Description %d", i, rand.Int63())

		var featureId int
		insertErr := db.QueryRow(`
			INSERT INTO features (name, description) 
			VALUES ($1, $2) 
			RETURNING id`, featureName, featureDescription).Scan(&featureId)
		if insertErr != nil {
			t.Fatal(insertErr)
		}

		_, linkErr := db.Exec(`
			INSERT INTO plan_features (plan_id, feature_id) 
			VALUES ($1, $2)`, planId, featureId)
		if linkErr != nil {
			t.Fatal(linkErr)
		}
	}
}
