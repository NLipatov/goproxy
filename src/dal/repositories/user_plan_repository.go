package repositories

import (
	"database/sql"
	"fmt"
	"goproxy/domain/aggregates"
	"log"
	"time"
)

const selectUserPlanById = "SELECT id, user_id, plan_id, valid_to, created_at FROM plans.public.user_plans WHERE id=$1"
const selectActiveUserPlans = "SELECT id, user_id, plan_id, valid_to, created_at FROM plans.public.user_plans WHERE valid_to > now()"
const selectUserPlans = "SELECT id, user_id, plan_id, valid_to, created_at FROM plans.public.user_plans"
const insertUserPlan = "INSERT INTO plans.public.user_plans (user_id, plan_id, valid_to) VALUES ($1, $2, $3) RETURNING id"
const updateUserPlan = "UPDATE plans.public.user_plans SET user_id=$1, plan_id=$2, valid_to=$3 WHERE id = $4 RETURNING id"
const deleteUserPlan = "DELETE FROM plans.public.user_plans WHERE id = $1"

type UserPlanRepository struct {
	db *sql.DB
}

func NewUserPlanRepository(db *sql.DB) *UserPlanRepository {
	return &UserPlanRepository{
		db: db,
	}
}

func (up *UserPlanRepository) GetUserActivePlans(userId int) ([]aggregates.UserPlan, error) {
	rows, err := up.db.Query(selectActiveUserPlans, userId)
	if err != nil {
		return nil, err
	}

	var userPlans []aggregates.UserPlan
	defer func(rows *sql.Rows) {
		closeErr := rows.Close()
		if closeErr != nil {
			log.Printf("failed to close rows: %v", closeErr)
		}
	}(rows)

	for rows.Next() {
		var id int
		var uId int
		var pId int
		var validTo time.Time
		var createdAt time.Time

		scanErr := rows.Scan(&id, &uId, &pId, &validTo, &createdAt)
		if scanErr != nil {
			return nil, scanErr
		}

		userPlan, validationErr := aggregates.NewUserPlan(id, uId, pId, validTo, createdAt)
		if validationErr != nil {
			return nil, validationErr
		}

		userPlans = append(userPlans, userPlan)
	}

	iterationErr := rows.Err()
	if iterationErr != nil {
		return nil, iterationErr
	}

	return userPlans, nil
}

func (up *UserPlanRepository) GetById(id int) (aggregates.UserPlan, error) {
	var userPlanId int
	var userId int
	var planId int
	var validTo time.Time
	var createdAt time.Time

	err := up.db.
		QueryRow(selectUserPlanById, id).
		Scan(&userPlanId, &userId, &planId, &validTo, &createdAt)

	if err != nil {
		return aggregates.UserPlan{}, err
	}

	userPlan, validationErr := aggregates.NewUserPlan(userPlanId, userId, planId, validTo, createdAt)
	if validationErr != nil {
		return aggregates.UserPlan{}, fmt.Errorf("validation error for user_plan %d: %s ", id, err)
	}

	return userPlan, nil
}

func (up *UserPlanRepository) Create(userPlan aggregates.UserPlan) (int, error) {
	var userPlanId int
	err := up.db.
		QueryRow(insertUserPlan, userPlan.UserId(), userPlan.PlanId(), userPlan.ValidTo()).
		Scan(&userPlanId)

	if err != nil {
		return 0, err
	}

	return userPlanId, nil
}

func (up *UserPlanRepository) Update(userPlan aggregates.UserPlan) error {
	result, err := up.db.
		Exec(updateUserPlan, userPlan.UserId(), userPlan.PlanId(), userPlan.ValidTo(), userPlan.Id())

	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil || affected == 0 {
		return fmt.Errorf("no rows updated for user plan id: %d", userPlan.Id())
	}

	return nil
}

func (up *UserPlanRepository) Delete(userPlan aggregates.UserPlan) error {
	result, err := up.db.Exec(deleteUserPlan, userPlan.Id())
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil || affected == 0 {
		return fmt.Errorf("no rows deleted for user_plan id: %d", userPlan.Id())
	}

	return nil
}
