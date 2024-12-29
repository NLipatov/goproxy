package events

type OutboxEvent struct {
	Id        int
	Payload   string
	Published bool
}

func NewOutboxEvent(id int, payload string, published bool) OutboxEvent {
	return OutboxEvent{
		Id:        id,
		Payload:   payload,
		Published: published,
	}
}
