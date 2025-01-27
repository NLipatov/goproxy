package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"goproxy/application"
	"goproxy/dal/cache_serialization"
	"goproxy/domain/dataobjects"
	"time"
)

const (
	planPriceCacheTtl = time.Hour * 12

	selectPlanPricesByPlanId = `
SELECT id, plan_id, currency, cents FROM plan_prices
WHERE plan_id = $1`

	selectPlanPricesById = `
SELECT id, plan_id, currency, cents FROM plan_prices
WHERE id = $1`

	insertPlanPrice = `
INSERT INTO plan_prices (plan_id, currency, cents) VALUES ($1, $2, $3) RETURNING id
`

	updatePlanPrice = `
UPDATE plan_prices 
SET plan_id = $1,
    currency = $2,
    cents = $3
WHERE id = $4
`

	deletePlanPrice = `
DELETE FROM plan_prices
WHERE id = $1
`
)

type PlanPriceRepository struct {
	db         *sql.DB
	cache      application.CacheWithTTL[[]cache_serialization.PriceDto]
	serializer cache_serialization.CacheSerializer[dataobjects.PlanPrice, cache_serialization.PriceDto]
}

func NewPlanPriceRepository(db *sql.DB, cache application.CacheWithTTL[[]cache_serialization.PriceDto]) application.PlanPriceRepository {
	return &PlanPriceRepository{
		db:         db,
		cache:      cache,
		serializer: cache_serialization.NewPriceCacheSerializer(),
	}
}

func (p *PlanPriceRepository) GetById(id int) (dataobjects.PlanPrice, error) {
	cacheKey := fmt.Sprintf("GetById_%d", id)
	cached, cachedErr := p.cache.Get(cacheKey)
	if cachedErr == nil {
		if len(cached) > 0 {
			return p.serializer.ToT(cached[0]), nil
		}
	}

	var priceId int
	var planId int
	var currency string
	var cents int64
	err := p.db.QueryRow(selectPlanPricesById, id).Scan(&planId, &priceId, &currency, &cents)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dataobjects.PlanPrice{}, fmt.Errorf("plan price not found")
		}
		return dataobjects.PlanPrice{}, fmt.Errorf("could not get plan prices by plan id %d: %w", id, err)
	}

	planPrice := dataobjects.NewPlanPrice(id, priceId, cents, currency)

	planPriceArray := make([]dataobjects.PlanPrice, 1)
	planPriceArray[0] = planPrice
	_ = p.cache.Set(cacheKey, p.serializer.ToDArray(planPriceArray))
	_ = p.cache.Expire(cacheKey, planPriceCacheTtl)

	return planPrice, nil
}

func (p *PlanPriceRepository) Create(entity dataobjects.PlanPrice) (int, error) {
	var id int
	err := p.db.QueryRow(insertPlanPrice, entity.PlanId(), entity.Currency(), entity.Cents()).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("could not create plan price: %w", err)
	}

	return id, nil
}

func (p *PlanPriceRepository) Update(entity dataobjects.PlanPrice) error {
	result, resultErr := p.db.Exec(updatePlanPrice, entity.PlanId(), entity.Currency(), entity.Cents(), entity.Id())
	if resultErr != nil {
		return fmt.Errorf("could not update plan price: %s", resultErr)
	}

	rowsAffectedErr := checkRowsAffected(result)
	if rowsAffectedErr != nil {
		return fmt.Errorf("could not update plan price: %s", rowsAffectedErr)
	}

	return nil

}

func (p *PlanPriceRepository) Delete(entity dataobjects.PlanPrice) error {
	result, resultErr := p.db.Exec(deletePlanPrice, entity.Id())
	if resultErr != nil {
		return fmt.Errorf("could not delete plan price: %s", resultErr)
	}

	rowsAffectedErr := checkRowsAffected(result)
	if rowsAffectedErr != nil {
		return fmt.Errorf("could not delete plan price: %s", rowsAffectedErr)
	}

	return nil
}

func (p *PlanPriceRepository) GetAllWithPlanId(planId int) ([]dataobjects.PlanPrice, error) {
	cacheKey := fmt.Sprintf("GetAllWithPlanId_%d", planId)
	cached, cachedErr := p.cache.Get(cacheKey)
	if cachedErr == nil {
		return p.serializer.ToTArray(cached), nil
	}

	rows, err := p.db.Query(selectPlanPricesByPlanId, planId)
	if err != nil {
		return nil, fmt.Errorf("failed to query plans: %v", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var planPrices []dataobjects.PlanPrice
	for rows.Next() {
		var id int
		var pId int
		var currency string
		var cents int64

		scanErr := rows.Scan(&id, &pId, &currency, &cents)
		if scanErr != nil {
			return nil, fmt.Errorf("failed to scan plan prices: %v", scanErr)
		}

		planPrice := dataobjects.NewPlanPrice(id, pId, cents, currency)
		planPrices = append(planPrices, planPrice)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("failed to read plan prices: %v", rows.Err())
	}

	_ = p.cache.Set(cacheKey, p.serializer.ToDArray(planPrices))
	_ = p.cache.Expire(cacheKey, planPriceCacheTtl)

	return planPrices, nil
}

func checkRowsAffected(result sql.Result) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not check rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("no rows affected")
	}
	return nil
}
