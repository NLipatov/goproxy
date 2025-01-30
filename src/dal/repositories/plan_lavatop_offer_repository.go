package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"goproxy/application/contracts"
	"goproxy/dal/cache_serialization"
	"goproxy/domain/dataobjects"
	"time"
)

const (
	cacheTtl = time.Hour

	UpdatePlanOffer = `
UPDATE plan_lavatop_offer
SET plan_id = $1, offer_id = $2
WHERE id = $3
`

	SelectOfferIdsByPlanId = `
SELECT id, plan_id, offer_id from plan_lavatop_offer
WHERE plan_id = $1
`
	InsertPlanIdAndOfferId = `
INSERT INTO plan_lavatop_offer (plan_id, offer_id)  
VALUES ($1, $2)
RETURNING id;
`

	DeletePlanOfferById = `
DELETE FROM plan_lavatop_offer
WHERE id = $1
`
)

type PlanLavatopOfferRepository struct {
	db                              *sql.DB
	cache                           contracts.CacheWithTTL[[]cache_serialization.PlanLavatopOfferDto]
	cachePlanLavatopOfferSerializer cache_serialization.CacheSerializer[dataobjects.PlanLavatopOffer, cache_serialization.PlanLavatopOfferDto]
}

func NewPlanLavatopOfferRepository(db *sql.DB,
	cache contracts.CacheWithTTL[[]cache_serialization.PlanLavatopOfferDto]) contracts.PlanOfferRepository {
	return &PlanLavatopOfferRepository{
		db:                              db,
		cache:                           cache,
		cachePlanLavatopOfferSerializer: cache_serialization.NewPlanLavatopOfferSerializer(),
	}
}

func (p *PlanLavatopOfferRepository) GetOffers(planId int) ([]dataobjects.PlanLavatopOffer, error) {
	planKey := p.planToCacheKey(planId)
	cached, cachedErr := p.cache.Get(planKey)
	if cachedErr == nil {
		return p.cachePlanLavatopOfferSerializer.ToTArray(cached), nil
	}

	var offers []dataobjects.PlanLavatopOffer
	rows, rowsErr := p.db.Query(SelectOfferIdsByPlanId, planId)
	if rowsErr != nil {
		return make([]dataobjects.PlanLavatopOffer, 0), rowsErr
	}

	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	for rows.Next() {
		var rId int
		var rPlanId int
		var rOfferId string

		scanErr := rows.Scan(&rId, &rPlanId, &rOfferId)
		if scanErr != nil {
			return make([]dataobjects.PlanLavatopOffer, 0), scanErr
		}

		offers = append(offers, dataobjects.NewPlanLavatopOffer(rId, rPlanId, rOfferId))
	}

	if rows.Err() != nil {
		return make([]dataobjects.PlanLavatopOffer, 0), rows.Err()
	}

	_ = p.cache.Set(planKey, p.cachePlanLavatopOfferSerializer.ToDArray(offers))
	_ = p.cache.Expire(planKey, cacheTtl)

	if offers == nil {
		return make([]dataobjects.PlanLavatopOffer, 0), nil
	}

	return offers, nil
}

func (p *PlanLavatopOfferRepository) Create(plo dataobjects.PlanLavatopOffer) (int, error) {
	var id int
	execErr := p.db.QueryRow(InsertPlanIdAndOfferId, plo.PlanId(), plo.OfferId()).Scan(&id)
	if execErr != nil {
		return 0, execErr
	}

	_ = p.cache.Expire(p.planToCacheKey(plo.PlanId()), 0)

	return id, nil
}

func (p *PlanLavatopOfferRepository) Update(plo dataobjects.PlanLavatopOffer) error {
	result, updateErr := p.db.Exec(UpdatePlanOffer, plo.PlanId(), plo.OfferId(), plo.Id())
	if updateErr != nil {
		return updateErr
	}
	rowsAffected, rowsAffectedErr := result.RowsAffected()
	if rowsAffectedErr != nil {
		return rowsAffectedErr
	}

	if rowsAffected != 1 {
		return errors.New("failed to update plan_lavatop_offer")
	}

	_ = p.cache.Expire(p.planToCacheKey(plo.PlanId()), 0)

	return nil
}

func (p *PlanLavatopOfferRepository) Delete(plo dataobjects.PlanLavatopOffer) error {
	result, execErr := p.db.Exec(DeletePlanOfferById, plo.Id())
	if execErr != nil {
		return execErr
	}

	if rowsAffected, err := result.RowsAffected(); err != nil || rowsAffected == 0 {
		return fmt.Errorf("0 rows affected")
	}

	_ = p.cache.Expire(p.planToCacheKey(plo.PlanId()), 0)

	return nil
}

func (p *PlanLavatopOfferRepository) planToCacheKey(planId int) string {
	return fmt.Sprintf("plan-to-lavatop-offers:plan_%d", planId)
}
