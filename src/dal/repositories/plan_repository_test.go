package repositories

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"goproxy/domain/aggregates"
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
		assertNoError(t, err, "Failed to load plan by ID")
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

		updatedPlan, _ := aggregates.NewPlan(planId, "Updated Plan", 2000000, 60)
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
}

func insertTestPlan(repo *PlanRepository, t *testing.T) int {
	name := fmt.Sprintf("Test Plan %d", time.Now().UTC().UnixNano())
	plan, err := aggregates.NewPlan(-1, name, 1000000, 30)
	assertNoError(t, err, "Failed to create test plan lavatopaggregates")

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
