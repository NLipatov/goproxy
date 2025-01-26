package events

type OrderPaidEvent struct {
	OrderId string `json:"order_id"`
}

func NewOrderPaidEvent(orderId string) OrderPaidEvent {
	return OrderPaidEvent{
		OrderId: orderId,
	}
}
