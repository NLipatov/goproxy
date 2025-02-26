package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"goproxy/application/contracts"
	"goproxy/dal/cache_serialization"
	"goproxy/domain/aggregates"
	"goproxy/domain/valueobjects"
	"time"
)

const (
	planRepositoryCacheTtl = time.Hour

	selectPlans = `
SELECT id, name, limit_bytes, duration_days, created_at
FROM plans
`

	selectPlansWithFeatures = `
SELECT
    plans.id AS plan_id,
    plans.name,
    plans.limit_bytes,
    plans.duration_days,
    features.name AS feature_name,
    features.description AS feature_description,
    plans.created_at
FROM plans
    LEFT JOIN plan_features ON plans.id = plan_features.plan_id
    LEFT JOIN features ON plan_features.feature_id = features.id;
`

	selectPlanByNameQuery = `
SELECT id, name, limit_bytes, duration_days, created_at 
FROM plans.public.plans 
WHERE name = $1`

	selectPlanByIdQuery = `
SELECT id, name, limit_bytes, duration_days, created_at 
FROM plans.public.plans 
WHERE id = $1`

	insertPlanQuery = `
INSERT INTO plans.public.plans (name, limit_bytes, duration_days) 
VALUES ($1, $2, $3) 
RETURNING id`

	updatePlanQuery = `
UPDATE plans.public.plans 
SET name=$1, limit_bytes=$2, duration_days=$3 
WHERE id = $4 
RETURNING id`

	deletePlanQuery = `
DELETE FROM plans.public.plans WHERE id = $1`

	GetPlanByIdWithFeaturesQuery = `
SELECT
    plans.id AS plan_id,
    plans.name,
    plans.limit_bytes,
    plans.duration_days,
    features.id AS feature_id,
    features.name AS feature_name,
    features.description AS feature_description,
    plans.created_at
FROM plans
         LEFT JOIN plan_features ON plans.id = plan_features.plan_id
         LEFT JOIN features ON plan_features.feature_id = features.id
WHERE plans.id = $1;
`

	GetPlanByNameWithFeaturesQuery = `
SELECT
    plans.id AS plan_id,
    plans.name,
    plans.limit_bytes,
    plans.duration_days,
    features.id AS feature_id,
    features.name AS feature_name,
    features.description AS feature_description,
    plans.created_at
FROM plans
         LEFT JOIN plan_features ON plans.id = plan_features.plan_id
         LEFT JOIN features ON plan_features.feature_id = features.id
WHERE plans.name = $1;
`
)

type PlanRepository struct {
	db                  *sql.DB
	cache               contracts.CacheWithTTL[[]cache_serialization.PlanDto]
	planCacheSerializer cache_serialization.CacheSerializer[aggregates.Plan, cache_serialization.PlanDto]
}

func NewPlansRepository(db *sql.DB, cache contracts.CacheWithTTL[[]cache_serialization.PlanDto]) contracts.PlanRepository {
	return &PlanRepository{
		db:                  db,
		cache:               cache,
		planCacheSerializer: cache_serialization.NewAggegatePlanCacheSerializer(),
	}
}

func (p *PlanRepository) GetAll() ([]aggregates.Plan, error) {
	cached, cachedErr := p.cache.Get(p.allPlansCacheKey())
	if cachedErr == nil {
		return p.planCacheSerializer.ToTArray(cached), nil
	}

	rows, err := p.db.Query(selectPlans)
	if err != nil {
		return nil, fmt.Errorf("failed to query plans: %v", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var plans []aggregates.Plan
	for rows.Next() {
		var id int
		var name string
		var limitBytes int64
		var durationDays int
		var createdAt time.Time

		err = rows.Scan(&id, &name, &limitBytes, &durationDays, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}

		plan, err := aggregates.NewPlan(id, name, limitBytes, durationDays, []valueobjects.PlanFeature{})
		if err != nil {
			fmt.Printf("plan validation err (invalid plan stored in db?): %s, plan id: %d\n", err, id)
			continue
		}
		plans = append(plans, plan)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("rows iteration error: %v", rows.Err())
	}

	_ = p.cache.Set(p.allPlansCacheKey(), p.planCacheSerializer.ToDArray(plans))
	_ = p.cache.Expire(p.allPlansCacheKey(), planRepositoryCacheTtl)

	return plans, nil
}

func (p *PlanRepository) GetAllWithFeatures() ([]aggregates.Plan, error) {
	cached, cachedErr := p.cache.Get(p.allPlansWithFeaturesCacheKey())
	if cachedErr == nil {
		return p.planCacheSerializer.ToTArray(cached), nil
	}

	rows, err := p.db.Query(selectPlansWithFeatures)
	if err != nil {
		return nil, fmt.Errorf("failed to query plans with features: %v", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	type planData struct {
		id         int
		name       string
		limitBytes int64
		duration   int
		createdAt  time.Time
		features   []valueobjects.PlanFeature
	}

	plansMap := make(map[int]*planData)

	for rows.Next() {
		var planId int
		var name string
		var limitBytes int64
		var durationDays int
		var createdAt time.Time
		var featureName sql.NullString
		var featureDescription sql.NullString

		err = rows.Scan(&planId, &name, &limitBytes, &durationDays, &featureName, &featureDescription, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}

		plan, exists := plansMap[planId]
		if !exists {
			plansMap[planId] = &planData{
				id:         planId,
				name:       name,
				limitBytes: limitBytes,
				duration:   durationDays,
				createdAt:  createdAt,
				features:   []valueobjects.PlanFeature{},
			}
			plan = plansMap[planId]
		}

		if featureName.Valid {
			feature := valueobjects.NewPlanFeature(planId, featureName.String, featureDescription.String)
			plan.features = append(plan.features, feature)
		}
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("rows iteration error: %v", rows.Err())
	}

	var plans []aggregates.Plan
	for _, data := range plansMap {
		plan, err := aggregates.NewPlan(data.id, data.name, data.limitBytes, data.duration, data.features)
		if err != nil {
			fmt.Printf("plan validation err (invalid plan data?): %s, plan id: %d\n", err, data.id)
			continue
		}
		plans = append(plans, plan)
	}

	_ = p.cache.Set(p.allPlansWithFeaturesCacheKey(), p.planCacheSerializer.ToDArray(plans))
	_ = p.cache.Expire(p.allPlansWithFeaturesCacheKey(), planRepositoryCacheTtl)

	return plans, nil
}

func (p *PlanRepository) GetByName(name string) (aggregates.Plan, error) {
	cached, cachedErr := p.cache.Get(p.allPlansWithFeaturesCacheKey())
	if cachedErr == nil && len(cached) == 1 {
		return p.planCacheSerializer.ToT(cached[0]), nil
	}

	plan, planErr := p.getPlan(selectPlanByNameQuery, name)
	if planErr != nil {
		return aggregates.Plan{}, planErr
	}

	planArr := make([]aggregates.Plan, 1)
	planArr[0] = plan
	_ = p.cache.Set(p.planNameToCacheKey(name), p.planCacheSerializer.ToDArray(planArr))
	_ = p.cache.Expire(p.planNameToCacheKey(name), planRepositoryCacheTtl)

	return plan, nil
}

func (p *PlanRepository) GetById(id int) (aggregates.Plan, error) {
	cached, cachedErr := p.cache.Get(p.allPlansWithFeaturesCacheKey())
	if cachedErr == nil && len(cached) == 1 {
		return p.planCacheSerializer.ToT(cached[0]), nil
	}

	plan, planErr := p.getPlan(selectPlanByIdQuery, id)
	if planErr != nil {
		return aggregates.Plan{}, planErr
	}

	planArr := make([]aggregates.Plan, 1)
	planArr[0] = plan
	_ = p.cache.Set(p.planIdToCacheKey(plan.Id()), p.planCacheSerializer.ToDArray(planArr))
	_ = p.cache.Expire(p.planIdToCacheKey(plan.Id()), planRepositoryCacheTtl)

	return plan, nil
}

func (p *PlanRepository) getPlan(query string, args ...interface{}) (aggregates.Plan, error) {
	var Id int
	var Name string
	var LimitBytes int64
	var DurationDays int
	var CreatedAt time.Time

	err := p.db.QueryRow(query, args...).Scan(&Id, &Name, &LimitBytes, &DurationDays, &CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return aggregates.Plan{}, fmt.Errorf("plan not found")
		}
		return aggregates.Plan{}, err
	}

	plan, err := aggregates.NewPlan(Id, Name, LimitBytes, DurationDays, make([]valueobjects.PlanFeature, 0))
	if err != nil {
		fmt.Printf("plan validation err (invalid plan stored in db?): %s, plan id: %d", err, Id)
	}

	return plan, nil
}

func (p *PlanRepository) Create(plan aggregates.Plan) (int, error) {
	var id int
	err := p.db.
		QueryRow(insertPlanQuery, plan.Name(), plan.LimitBytes(), plan.DurationDays()).
		Scan(&id)
	return id, err
}

func (p *PlanRepository) Update(plan aggregates.Plan) error {
	_, err := p.db.
		Exec(updatePlanQuery, plan.Name(), plan.LimitBytes(), plan.DurationDays(), plan.Id())

	if err != nil {
		return fmt.Errorf("could not update plan: %v", err)
	}
	return nil
}

func (p *PlanRepository) Delete(plan aggregates.Plan) error {
	result, err := p.db.Exec(deletePlanQuery, plan.Id())
	if err != nil {
		return fmt.Errorf("could not delete plan: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil || affected == 0 {
		return fmt.Errorf("no rows affected")
	}
	return nil
}

func (p *PlanRepository) GetByIdWithFeatures(id int) (aggregates.Plan, error) {
	return p.getPlanWithFeatures(GetPlanByIdWithFeaturesQuery, id)
}

func (p *PlanRepository) GetByNameWithFeatures(name string) (aggregates.Plan, error) {
	return p.getPlanWithFeatures(GetPlanByNameWithFeaturesQuery, name)
}

func (p *PlanRepository) getPlanWithFeatures(query string, arg interface{}) (aggregates.Plan, error) {
	var planId int
	var name string
	var limitBytes int64
	var durationDays int
	var createdAt time.Time

	var features []valueobjects.PlanFeature

	rows, err := p.db.Query(query, arg)
	if err != nil {
		return aggregates.Plan{}, fmt.Errorf("failed to query plan with features: %v", err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	for rows.Next() {
		var featureId sql.NullInt64
		var featureName sql.NullString
		var featureDescription sql.NullString

		err = rows.Scan(&planId, &name, &limitBytes, &durationDays, &featureId, &featureName, &featureDescription, &createdAt)
		if err != nil {
			return aggregates.Plan{}, fmt.Errorf("failed to scan row (query: %s, arg: %v): %w", query, arg, err)
		}

		if featureDescription.Valid {
			if featureId.Valid {
				pId := int(featureId.Int64)
				features = append(features, valueobjects.NewPlanFeature(pId, featureName.String, featureDescription.String))
			}
		}
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return aggregates.Plan{}, fmt.Errorf("rows iteration error: %v", rowsErr)
	}

	return aggregates.NewPlan(planId, name, limitBytes, durationDays, features)
}

func (p *PlanRepository) planIdToCacheKey(id int) string {
	return fmt.Sprintf("plan_repository:plan:%d", id)
}

func (p *PlanRepository) planNameToCacheKey(name string) string {
	return fmt.Sprintf("plan_repository:plan:%s", name)
}

func (p *PlanRepository) allPlansCacheKey() string {
	return "plan_repository:plan:*"
}

func (p *PlanRepository) allPlansWithFeaturesCacheKey() string {
	return "plan_repository:plan_with_features:*"
}
