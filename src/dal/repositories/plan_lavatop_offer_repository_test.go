package repositories

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"goproxy/application/contracts"
	"goproxy/dal/cache_serialization"
	"goproxy/dal/repositories/mocks"
	"goproxy/domain/dataobjects"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlanLavatopOfferRepository(t *testing.T) {
	setEnvErr := os.Setenv("DB_DATABASE", "plans")
	if setEnvErr != nil {
		t.Fatal(setEnvErr)
	}

	defer func() {
		_ = os.Unsetenv("DB_DATABASE")
	}()

	db, cleanup := prepareCockroachDB(t)
	defer cleanup()
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	planOfferRepoCache := mocks.NewMockCacheWithTTL[[]cache_serialization.PlanLavatopOfferDto]()
	planRepoCache := mocks.NewMockCacheWithTTL[[]cache_serialization.PlanDto]()
	planOfferRepo := NewPlanLavatopOfferRepository(db, planOfferRepoCache)
	planRepo := NewPlansRepository(db, planRepoCache)

	t.Run("GetOffers", func(t *testing.T) {
		ploCount := 5
		planId := insertTestPlan(planRepo, t)

		expected := insertTestPlanOffers(planOfferRepo, planId, ploCount, t)

		actual, err := planOfferRepo.GetOffers(planId)

		var matched int
		for _, e := range expected {
			for _, a := range actual {
				if e.Id() == a.Id() &&
					e.OfferId() == a.OfferId() &&
					e.PlanId() == a.PlanId() {
					matched++
				}
			}
		}

		assert.Nil(t, err)
		assert.Equal(t, matched, ploCount)
		assert.Equal(t, ploCount, len(actual))
	})

	t.Run("Create", func(t *testing.T) {
		planId := insertTestPlan(planRepo, t)
		offerId := fmt.Sprintf("%s", uuid.New())

		ploId, createErr := planOfferRepo.Create(dataobjects.NewPlanLavatopOffer(-1, planId, offerId))
		offers, err := planOfferRepo.GetOffers(planId)

		assert.Nil(t, err)
		assert.Nil(t, createErr)
		assert.Equal(t, offerId, offers[0].OfferId())
		assert.Equal(t, planId, offers[0].PlanId())
		assert.Equal(t, ploId, offers[0].Id())
	})

	t.Run("Update", func(t *testing.T) {
		planId := insertTestPlan(planRepo, t)
		offerId := fmt.Sprintf("%s", uuid.New())
		updatedOfferId := fmt.Sprintf("%s", uuid.New())

		ploId, createErr := planOfferRepo.Create(dataobjects.NewPlanLavatopOffer(-1, planId, offerId))

		updateErr := planOfferRepo.Update(dataobjects.NewPlanLavatopOffer(ploId, planId, updatedOfferId))
		offers, err := planOfferRepo.GetOffers(planId)

		assert.Nil(t, err)
		assert.Nil(t, createErr)
		assert.Nil(t, updateErr)
		assert.Equal(t, updatedOfferId, offers[0].OfferId())
		assert.NotEqual(t, ploId, offers[0].OfferId())
		assert.Equal(t, planId, offers[0].PlanId())
		assert.Equal(t, ploId, offers[0].Id())
	})

	t.Run("Delete", func(t *testing.T) {
		planId := insertTestPlan(planRepo, t)
		offerId := fmt.Sprintf("%s", uuid.New())

		ploId, createErr := planOfferRepo.Create(dataobjects.NewPlanLavatopOffer(-1, planId, offerId))

		deleteErr := planOfferRepo.Delete(dataobjects.NewPlanLavatopOffer(ploId, planId, offerId))

		offers, offersErr := planOfferRepo.GetOffers(planId)

		assert.Nil(t, createErr)
		assert.Nil(t, deleteErr)
		assert.Nil(t, offersErr)
		assert.Empty(t, offers)
	})
}

func insertTestPlanOffers(repo contracts.PlanOfferRepository, planId, ploCount int, t *testing.T) []dataobjects.PlanLavatopOffer {
	insertedPlos := make([]dataobjects.PlanLavatopOffer, ploCount)
	for i := 0; i < ploCount; i++ {
		offerId := fmt.Sprintf("%s", uuid.New())

		ploId, createPloErr := repo.Create(dataobjects.NewPlanLavatopOffer(-1, planId, offerId))
		if createPloErr != nil {
			t.Fatal(createPloErr)
		}

		insertedPlos[i] = dataobjects.NewPlanLavatopOffer(ploId, planId, offerId)
	}

	return insertedPlos
}
