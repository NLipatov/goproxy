package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"goproxy/application/contracts"
	"goproxy/domain/dataobjects"
	"goproxy/domain/valueobjects"
)

const (
	selectOrderById = `
SELECT id, email, plan_id, status
FROM orders
WHERE id = $1`

	selectOrdersByPlanIdAndEmail = `
SELECT id, email, plan_id, status
FROM orders
WHERE plan_id = $1 AND email = $2`

	insertOrder = `
INSERT INTO orders (email, plan_id, status)
VALUES ($1, $2, $3) RETURNING id`

	updateOrder = `
UPDATE orders
SET email = $1,
    plan_id = $2,
    status = $3
WHERE id = $4`

	deleteOrder = `
DELETE FROM orders
WHERE id = $1`
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) contracts.OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) GetByPlanIdAndEmail(planId int, email valueobjects.Email) ([]dataobjects.Order, error) {
	rows, err := r.db.Query(selectOrdersByPlanIdAndEmail, planId, email.String())
	if err != nil {
		return nil, fmt.Errorf("could not get orders by plan_id %d and email %s: %w", planId, email.String(), err)
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	var orders []dataobjects.Order
	for rows.Next() {
		var oId int
		var oEmail string
		var oPlanId int
		var status string

		err = rows.Scan(&oId, &oEmail, &oPlanId, &status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order status: %w", err)
		}

		emailVO, emailErr := valueobjects.ParseEmailFromString(oEmail)
		if emailErr != nil {
			return nil, fmt.Errorf("order %d has invalid email: %s", oId, emailErr)
		}
		orders = append(orders, dataobjects.NewOrder(oId, emailVO, oPlanId, valueobjects.NewOrderStatus(status)))
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error reading rows: %w", rows.Err())
	}

	return orders, nil
}

func (r *OrderRepository) GetById(id int) (dataobjects.Order, error) {
	var orderId int
	var email string
	var planId int
	var status string

	err := r.db.QueryRow(selectOrderById, id).Scan(&orderId, &email, &planId, &status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dataobjects.Order{}, fmt.Errorf("order not found")
		}
		return dataobjects.Order{}, fmt.Errorf("could not get order by id %d: %w", id, err)
	}

	emailVO, emailErr := valueobjects.ParseEmailFromString(email)
	if emailErr != nil {
		return dataobjects.Order{}, fmt.Errorf("invalid email in database: %w", emailErr)
	}

	orderStatus := valueobjects.NewOrderStatus(status)

	return dataobjects.NewOrder(orderId, emailVO, planId, orderStatus), nil
}

func (r *OrderRepository) Create(entity dataobjects.Order) (int, error) {
	var id int
	err := r.db.QueryRow(insertOrder, entity.Email(), entity.PlanId(), entity.Status().String()).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("could not create order: %w", err)
	}

	return id, nil
}

func (r *OrderRepository) Update(entity dataobjects.Order) error {
	result, err := r.db.Exec(updateOrder, entity.Email(), entity.PlanId(), entity.Status().String(), entity.Id())
	if err != nil {
		return fmt.Errorf("could not update order: %w", err)
	}

	return NewSqlResult(result).checkRowsAffected()
}

func (r *OrderRepository) Delete(entity dataobjects.Order) error {
	result, err := r.db.Exec(deleteOrder, entity.Id())
	if err != nil {
		return fmt.Errorf("could not delete order: %w", err)
	}

	return NewSqlResult(result).checkRowsAffected()
}
