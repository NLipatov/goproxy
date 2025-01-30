package valueobjects

import (
	"fmt"
)

type OrderStatus struct {
	status string
}

func NewOrderStatus(status string) OrderStatus {
	return OrderStatus{
		status: status,
	}
}

func (o OrderStatus) String() string {
	return fmt.Sprintf("%s", o.status)
}
