package repositories

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
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

	db, cleanup := prepareDb(t)
	defer cleanup()
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	repo := NewPlansRepository(db)

	t.Run("GetByName", func(t *testing.T) {
		planId := insertTestPlan(repo, t)
		plan, err := repo.GetById(planId)
		assertNoError(t, err, "Failed to load plan by Id")
		loadedPlan, err := repo.GetByName(plan.Name())
		assertNoError(t, err, "Failed to load plan by Name")
		assertPlansEqual(t, plan, loadedPlan)
	})

	t.Run("GetById", func(t *testing.T) {
		planId := insertTestPlan(repo, t)
		plan, err := repo.GetById(planId)
		assertNoError(t, err, "Failed to load plan by Id")
		loadedPlan, err := repo.GetById(plan.Id())
		assertNoError(t, err, "Failed to load plan by Id")
		assertPlansEqual(t, plan, loadedPlan)
	})

	t.Run("Create", func(t *testing.T) {
		planId := insertTestPlan(repo, t)
		plan, err := repo.GetById(planId)
		assertNoError(t, err, "Failed to load plan by Id")
		loadedPlan, err := repo.GetByName(plan.Name())
		assertNoError(t, err, "Failed to load inserted plan")
		assertPlansEqual(t, plan, loadedPlan)
	})

	t.Run("Update", func(t *testing.T) {
		planId := insertTestPlan(repo, t)
		plan, err := repo.GetById(planId)
		assertNoError(t, err, "Failed to load plan by Id")

		updatedPlan, _ := aggregates.NewPlan(planId, "Updated Plan", 2000000, 60,
			make([]valueobjects.PlanFeature, 0))
		assertNoError(t, repo.Update(updatedPlan), "Failed to update plan")

		loadedPlan, err := repo.GetById(plan.Id())
		assertNoError(t, err, "Failed to load updated plan")
		assertPlansNotEqual(t, plan, loadedPlan)
	})

	t.Run("Delete", func(t *testing.T) {
		planId := insertTestPlan(repo, t)
		plan, err := repo.GetById(planId)
		assertNoError(t, err, "Failed to load plan by Id")

		assertNoError(t, repo.Delete(plan), "Failed to delete plan")

		_, err = repo.GetById(plan.Id())
		if err == nil {
			t.Error("Expected error when loading deleted plan")
		}
	})

	t.Run("GetByIdWithFeatures", func(t *testing.T) {
		planId := insertTestPlanWithFeatures(repo, t)
		plan, err := repo.GetByIdWithFeatures(planId)
		assertNoError(t, err, "Failed to load plan with features by Id")

		if len(plan.Features()) == 0 {
			t.Errorf("Expected features for plan with id %d, but got none", planId)
		}
	})

	t.Run("GetByNameWithFeatures", func(t *testing.T) {
		planId := insertTestPlanWithFeatures(repo, t)
		plan, err := repo.GetByIdWithFeatures(planId)
		assertNoError(t, err, "Failed to load plan with features by Id")

		loadedPlan, err := repo.GetByNameWithFeatures(plan.Name())
		assertNoError(t, err, "Failed to load plan with features by Name")
		assertPlansEqual(t, plan, loadedPlan)

		if len(loadedPlan.Features()) == 0 {
			t.Errorf("Expected features for plan with name %s, but got none", plan.Name())
		}
	})

}

func insertTestPlan(repo *PlanRepository, t *testing.T) int {
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

func insertTestPlanWithFeatures(repo *PlanRepository, t *testing.T) int {
	name := fmt.Sprintf("Test Plan With Features %d", time.Now().UTC().UnixNano())
	plan, err := aggregates.NewPlan(-1, name, time.Now().UnixMilli(), time.Now().Day(), make([]valueobjects.PlanFeature, 0))
	assertNoError(t, err, "Failed to create test plan")

	planId, err := repo.Create(plan)
	assertNoError(t, err, "Failed to insert test plan")

	addFeaturesToPlan(repo, t, planId, rand.Intn(10))

	return planId
}

func addFeaturesToPlan(repo *PlanRepository, t *testing.T, planId int, featureCount int) {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	db := repo.db

	for i := 0; i < featureCount; i++ {
		featureName := fmt.Sprintf("Feature %d %d", i, rand.Int63())
		featureDescription := fmt.Sprintf("Feature %d Description %d", i, rand.Int63())

		var featureId int
		insertErr := db.QueryRow(`
			INSERT INTO features (name, description) 
			VALUES ($1, $2) 
			RETURNING id`, featureName, featureDescription).Scan(&featureId)
		assertNoError(t, insertErr, "Failed to insert feature")

		_, linkErr := db.Exec(`
			INSERT INTO plan_features (plan_id, feature_id) 
			VALUES ($1, $2)`, planId, featureId)
		assertNoError(t, linkErr, "Failed to link feature to plan")
	}
}
