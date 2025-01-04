package repositories

import "database/sql"

const selectUserPlanById = "SELECT id, user_id, plan_id, valid_to, created_at FROM plans.public.user_plans WHERE id = $1"
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
