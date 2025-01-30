package dataobjects

import "goproxy/domain/valueobjects"

type Order struct {
	id     int
	email  valueobjects.Email
	planId int
	status valueobjects.OrderStatus
}

func NewOrder(id int, email valueobjects.Email, planId int, status valueobjects.OrderStatus) Order {
	return Order{
		id:     id,
		email:  email,
		planId: planId,
		status: status,
	}
}

func (o *Order) Id() int {
	return o.id
}

func (o *Order) Email() string {
	return o.email.String()
}

func (o *Order) PlanId() int {
	return o.planId
}

func (o *Order) Status() valueobjects.OrderStatus {
	return o.status
}
