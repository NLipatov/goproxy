package repositories

import (
	"database/sql"
	"fmt"
	"github.com/stretchr/testify/assert"
	"goproxy/dal/cache_serialization"
	"goproxy/dal/repositories/mocks"
	"goproxy/domain/dataobjects"
	"os"
	"reflect"
	"testing"

	_ "github.com/lib/pq"
)

func TestPlanPriceRepository(t *testing.T) {
	setEnvErr := os.Setenv("DB_DATABASE", "billing")
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

	planPriceCache := mocks.NewMockCacheWithTTL[[]cache_serialization.PriceDto]()
	planPriceRepository := NewPlanPriceRepository(db, planPriceCache)

	t.Run("GetById", func(t *testing.T) {
		expectedPlanId := 1
		expectedCents := int64(1000)
		expectedCurrency := "USD"
		priceId := insertTestPlanPrice(t, db, expectedPlanId, expectedCurrency, expectedCents)

		price, priceErr := planPriceRepository.GetById(priceId)

		assert.Nil(t, priceErr)
		assert.NotNil(t, price)
		assert.Equal(t, expectedPlanId, price.PlanId())
		assert.Equal(t, expectedCents, price.Cents())
		assert.Equal(t, expectedCurrency, price.Currency())
	})

	t.Run("GetById_CacheSetCalled", func(t *testing.T) {
		priceId := insertTestPlanPrice(t, db, 5, "RUB", 200_000)

		_, _ = planPriceRepository.GetById(priceId)
		cached, cachedErr := planPriceCache.Get(fmt.Sprintf("GetById_%d", priceId))

		assert.Nil(t, cachedErr)
		assert.Equal(t, cached[0].Cents, int64(200_000))
		assert.Equal(t, cached[0].Currency, "RUB")
	})

	t.Run("GetById_CacheGetCalled", func(t *testing.T) {
		expectedId := 100
		expectedPlanId := 200
		expectedCents := int64(3000)
		expectedCurrency := "DKK"
		setErr := planPriceCache.Set(fmt.Sprintf("GetById_%d", expectedId), []cache_serialization.PriceDto{
			{
				Id:       expectedId,
				PlanId:   expectedPlanId,
				Cents:    expectedCents,
				Currency: expectedCurrency,
			}})
		if setErr != nil {
			t.Fatal(setErr)
		}

		price, priceErr := planPriceRepository.GetById(100)
		if priceErr != nil {
			t.Fatal(priceErr)
		}

		assert.NotNil(t, price)
		assert.Equal(t, price.Id(), expectedId)
		assert.Equal(t, price.PlanId(), expectedPlanId)
		assert.Equal(t, price.Cents(), expectedCents)
		assert.Equal(t, price.Currency(), expectedCurrency)
	})

	t.Run("Create", func(t *testing.T) {
		expectedPlanId := 2
		expectedCents := int64(10_000)
		expectedCurrency := "EUR"

		expectedPrice := dataobjects.NewPlanPrice(-1, expectedPlanId, expectedCents, expectedCurrency)
		expectedPriceId, priceErr := planPriceRepository.Create(expectedPrice)
		if priceErr != nil {
			t.Fatal(priceErr)
		}

		actualPrice, actualPriceErr := planPriceRepository.GetById(expectedPriceId)
		if actualPriceErr != nil {
			t.Fatal(actualPriceErr)
		}

		assert.Equal(t, expectedPriceId, actualPrice.Id())
		assert.Equal(t, expectedPlanId, actualPrice.PlanId())
		assert.Equal(t, expectedCents, actualPrice.Cents())
		assert.Equal(t, expectedCurrency, actualPrice.Currency())
	})

	t.Run("Update", func(t *testing.T) {
		planId := 2
		cents := int64(10_000)
		currency := "EUR"

		price := dataobjects.NewPlanPrice(-1, planId, cents, currency)
		priceId, priceErr := planPriceRepository.Create(price)
		if priceErr != nil {
			t.Fatal(priceErr)
		}

		newPlanId := 3
		newCents := int64(80_000)
		newCurrency := "AUD"
		updatedPricePlan := dataobjects.NewPlanPrice(priceId, newPlanId, newCents, newCurrency)
		updatedPriceErr := planPriceRepository.Update(updatedPricePlan)
		if updatedPriceErr != nil {
			t.Fatal(updatedPriceErr)
		}

		updatedPrice, actualPriceErr := planPriceRepository.GetById(priceId)
		if actualPriceErr != nil {
			t.Fatal(actualPriceErr)
		}

		assert.Equal(t, priceId, updatedPrice.Id())
		assert.Equal(t, newPlanId, updatedPrice.PlanId())
		assert.Equal(t, newCents, updatedPrice.Cents())
		assert.Equal(t, newCurrency, updatedPrice.Currency())
	})

	t.Run("Delete", func(t *testing.T) {
		priceId := insertTestPlanPrice(t, db, 3, "PLN", 150_000)
		price, priceErr := planPriceRepository.GetById(priceId)
		if priceErr != nil {
			t.Fatal(priceErr)
		}

		deleteErr := planPriceRepository.Delete(price)
		if deleteErr != nil {
			t.Fatal(deleteErr)
		}

		price, priceErr = planPriceRepository.GetById(priceId)

		assert.NotNil(t, priceErr)
		assert.Contains(t, priceErr.Error(), "plan price not found")
	})

	t.Run("GetAllWithPlanId", func(t *testing.T) {
		planId := 4

		fPlanId := insertTestPlanPrice(t, db, planId, "EUR", 150_000)
		fPlan := dataobjects.NewPlanPrice(fPlanId, planId, 150_000, "EUR")

		sPlanId := insertTestPlanPrice(t, db, planId, "USD", 150_000)
		sPlan := dataobjects.NewPlanPrice(sPlanId, planId, 150_000, "USD")

		tPlanId := insertTestPlanPrice(t, db, planId, "AUD", 150_000)
		tPlan := dataobjects.NewPlanPrice(tPlanId, planId, 150_000, "AUD")

		foPlanId := insertTestPlanPrice(t, db, planId, "BYN", 150_000)
		foPlan := dataobjects.NewPlanPrice(foPlanId, planId, 150_000, "BYN")

		prices, pricesErr := planPriceRepository.GetAllWithPlanId(planId)
		if pricesErr != nil {
			t.Fatal(pricesErr)
		}

		for _, price := range prices {
			switch price.Id() {
			case fPlanId:
				assert.True(t, reflect.DeepEqual(fPlan, price))
			case sPlanId:
				assert.True(t, reflect.DeepEqual(sPlan, price))
			case tPlanId:
				assert.True(t, reflect.DeepEqual(tPlan, price))
			case foPlanId:
				assert.True(t, reflect.DeepEqual(foPlan, price))
			default:
				t.Fatalf("unexpected plan id: %d", price.PlanId())
			}
		}
	})
}

func insertTestPlanPrice(t *testing.T, db *sql.DB, planId int, currency string, cents int64) int {
	var id int
	resultErr := db.
		QueryRow(insertPlanPrice, planId, currency, cents).
		Scan(&id)

	if resultErr != nil {
		t.Fatalf("Failed to insert plan: %v", resultErr)
	}

	return id
}
