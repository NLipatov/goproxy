package repositories

import (
	"database/sql"
	"fmt"
	"goproxy/application"
)

const (
	SelectOfferIdsByPlanId = `
SELECT offer_id from plan_lavatop_offer
WHERE plan_id = $1
`
	InsertPlanIdAndOfferId = `
INSERT INTO plan_lavatop_offer (plan_id, offer_id)  
VALUES ($1, $2)
`

	DeleteByPlanIdAndOfferId = `
DELETE FROM plan_lavatop_offer
WHERE plan_id = $1 AND offer_id = $2
`
)

type PlanLavatopOfferRepository struct {
	db    *sql.DB
	cache application.CacheWithTTL[[]string]
}

func NewPlanLavatopOfferRepository(db *sql.DB, cache application.CacheWithTTL[[]string]) PlanLavatopOfferRepository {
	return PlanLavatopOfferRepository{
		db:    db,
		cache: cache,
	}
}

func (p *PlanLavatopOfferRepository) GetOfferIds(planId int) ([]string, error) {
	planKey := p.planToCacheKey(planId)
	cached, cachedErr := p.cache.Get(planKey)
	if cachedErr == nil {
		return cached, nil
	}

	var offerIds []string
	rows, rowsErr := p.db.Query(SelectOfferIdsByPlanId, planId)
	if rowsErr != nil {
		return offerIds, rowsErr
	}

	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	for rows.Next() {
		var offerId string

		scanErr := rows.Scan(&offerId)
		if scanErr != nil {
			return offerIds, scanErr
		}

		offerIds = append(offerIds, offerId)
	}

	if rows.Err() != nil {
		return offerIds, rows.Err()
	}

	_ = p.cache.Set(planKey, offerIds)

	return offerIds, nil
}

func (p *PlanLavatopOfferRepository) AddOffer(planId int, offerId string) (int64, error) {
	result, execErr := p.db.Exec(InsertPlanIdAndOfferId, planId, offerId)
	if execErr != nil {
		return 0, execErr
	}

	rowsAffected, rowsAffectedErr := result.RowsAffected()
	if rowsAffectedErr != nil {
		return 0, rowsAffectedErr
	}

	planKey := p.planToCacheKey(planId)
	cached, cachedErr := p.cache.Get(planKey)
	if cachedErr == nil {
		cached = append(cached, offerId)
		_ = p.cache.Set(planKey, cached)
	}

	return rowsAffected, nil
}

func (p *PlanLavatopOfferRepository) RemoveOffer(planId int, offerId string) (int64, error) {
	result, execErr := p.db.Exec(DeleteByPlanIdAndOfferId, planId, offerId)
	if execErr != nil {
		return 0, execErr
	}

	rowsAffected, rowsAffectedErr := result.RowsAffected()
	if rowsAffectedErr != nil {
		return rowsAffected, rowsAffectedErr
	}

	planKey := p.planToCacheKey(planId)
	cached, cachedErr := p.cache.Get(planKey)
	if cachedErr == nil {
		for i, id := range cached {
			if id == offerId {
				cached = append(cached[:i], cached[i+1:]...)
				break
			}
		}
		_ = p.cache.Set(planKey, cached)
	}

	return rowsAffected, nil
}

func (p *PlanLavatopOfferRepository) planToCacheKey(planId int) string {
	return fmt.Sprintf("plan-to-lavatop-offers:plan_%d", planId)
}
