package repositories

import (
	"database/sql"
	"goproxy/dal/repositories/mocks"
	"os"
	"testing"

	"github.com/google/uuid"
)

func TestPlanLavatopOfferRepository(t *testing.T) {
	setEnvErr := os.Setenv("DB_DATABASE", "plans")
	if setEnvErr != nil {
		t.Fatal(setEnvErr)
	}

	db, cleanup := prepareDb(t)
	defer cleanup()
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	cache := mocks.NewMockCacheWithTTL[[]string]()
	repo := NewPlanLavatopOfferRepository(db, cache)
	plansRepo := NewPlansRepository(db)

	t.Run("GetOfferIds", func(t *testing.T) {
		planId := insertTestPlan(plansRepo, t)
		offerIds := generateUUIDs(3)
		for _, offerId := range offerIds {
			rowsAffected, err := repo.AddOffer(planId, offerId)
			if err != nil {
				t.Fatal(err)
			}
			if rowsAffected == 0 {
				t.Fatal("expected to get a row affected, but got none")
			}
		}

		offers, err := repo.GetOfferIds(planId)
		assertNoError(t, err, "Failed to get offer IDs")
		if len(offers) != 3 {
			t.Errorf("Expected 3 offers, got %d", len(offers))
		}

		if !equalSlices(offerIds, offers) {
			t.Errorf("Expected offers %v, got %v", offerIds, offers)
		}

	})

	t.Run("AddOffer", func(t *testing.T) {
		planId := insertTestPlan(plansRepo, t)
		newOffer := uuid.New().String()
		rowsAffected, err := repo.AddOffer(planId, newOffer)
		assertNoError(t, err, "Failed to add offer")
		if rowsAffected != 1 {
			t.Errorf("Expected 1 row affected, got %d", rowsAffected)
		}

		offers, err := repo.GetOfferIds(planId)
		assertNoError(t, err, "Failed to get offer IDs after adding offer")
		if len(offers) != 1 || offers[0] != newOffer {
			t.Errorf("Expected %s in offers, got %v", newOffer, offers)
		}
	})

	t.Run("RemoveOffer", func(t *testing.T) {
		offers := generateUUIDs(1)
		planId := insertTestPlan(plansRepo, t)
		addOfferRowsAffected, addOfferErr := repo.AddOffer(planId, offers[0])
		assertNoError(t, addOfferErr, "Failed to add offer")

		if addOfferRowsAffected != 1 {
			t.Errorf("Expected 1 row affected, got %d", addOfferRowsAffected)
		}

		RemoveOfferRowsAffected, RemoveOfferErr := repo.RemoveOffer(planId, offers[0])
		assertNoError(t, RemoveOfferErr, "Failed to remove offer")
		if RemoveOfferRowsAffected != 1 {
			t.Errorf("Expected 1 row affected, got %d", RemoveOfferRowsAffected)
		}

		offersAfterRemoval, err := repo.GetOfferIds(planId)
		assertNoError(t, err, "Failed to get offer IDs after removing offer")
		if len(offersAfterRemoval) != 0 {
			t.Errorf("Expected no offers, got %v", offersAfterRemoval)
		}
	})
}

func generateUUIDs(count int) []string {
	uuids := make([]string, count)
	for i := 0; i < count; i++ {
		uuids[i] = uuid.New().String()
	}
	return uuids
}
