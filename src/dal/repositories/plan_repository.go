package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"goproxy/domain/aggregates"
	"goproxy/domain/valueobjects"
	"time"
)

const selectPlanByNameQuery = `
SELECT id, name, limit_bytes, duration_days, created_at 
FROM plans.public.plans 
WHERE name = $1`

const selectPlanByIdQuery = `
SELECT id, name, limit_bytes, duration_days, created_at 
FROM plans.public.plans 
WHERE id = $1`

const insertPlanQuery = `
INSERT INTO plans.public.plans (name, limit_bytes, duration_days) 
VALUES ($1, $2, $3) 
RETURNING id`

const updatePlanQuery = `
UPDATE plans.public.plans 
SET name=$1, limit_bytes=$2, duration_days=$3 
WHERE id = $4 
RETURNING id`

const deletePlanQuery = `
DELETE FROM plans.public.plans WHERE id = $1`

const GetPlanByIdWithFeaturesQuery = `
SELECT
    plans.id AS plan_id,
    plans.name,
    plans.limit_bytes,
    plans.duration_days,
    features.id AS feature_id,
    features.description AS feature_description,
    plans.created_at
FROM plans
         LEFT JOIN plan_features ON plans.id = plan_features.plan_id
         LEFT JOIN features ON plan_features.feature_id = features.id
WHERE plans.id = $1;
`

const GetPlanByNameWithFeaturesQuery = `
SELECT
    plans.id AS plan_id,
    plans.name,
    plans.limit_bytes,
    plans.duration_days,
    features.id AS feature_id,
    features.description AS feature_description,
    plans.created_at
FROM plans
         LEFT JOIN plan_features ON plans.id = plan_features.plan_id
         LEFT JOIN features ON plan_features.feature_id = features.id
WHERE plans.name = $1;
`

type PlanRepository struct {
	db *sql.DB
}

func NewPlansRepository(db *sql.DB) *PlanRepository {
	return &PlanRepository{
		db: db,
	}
}

func (p *PlanRepository) GetByName(name string) (aggregates.Plan, error) {
	return p.getPlan(selectPlanByNameQuery, name)
}

func (p *PlanRepository) GetById(id int) (aggregates.Plan, error) {
	return p.getPlan(selectPlanByIdQuery, id)
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
		var featureDescription sql.NullString

		err = rows.Scan(&planId, &name, &limitBytes, &durationDays, &featureId, &featureDescription, &createdAt)
		if err != nil {
			return aggregates.Plan{}, fmt.Errorf("failed to scan row (query: %s, arg: %v): %w", query, arg, err)
		}

		if featureDescription.Valid {
			if featureId.Valid {
				pId := int(featureId.Int64)
				features = append(features, valueobjects.NewPlanFeature(pId, featureDescription.String))
			}
		}
	}

	if rowsErr := rows.Err(); rowsErr != nil {
		return aggregates.Plan{}, fmt.Errorf("rows iteration error: %v", rowsErr)
	}

	return aggregates.NewPlan(planId, name, limitBytes, durationDays, features)
}
