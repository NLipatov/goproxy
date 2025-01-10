package lavatopvalueobjects

import (
	"errors"
	"strings"
)

type Status int

const (
	NEW Status = iota
	INPROGRESS
	COMPLETED
	FAILED
	CANCELLED
	SUBSCRIPTIONACTIVE
	SUBSCRIPTIONEXPIRED
	SUBSCRIPTIONCANCELLED
	SUBSCRIPTIONFAILED
)

func (i Status) String() string {
	switch i {
	case NEW:
		return "new"
	case INPROGRESS:
		return "in-progress"
	case COMPLETED:
		return "completed"
	case FAILED:
		return "failed"
	case CANCELLED:
		return "cancelled"
	case SUBSCRIPTIONACTIVE:
		return "subscription-active"
	case SUBSCRIPTIONEXPIRED:
		return "subscription-expired"
	case SUBSCRIPTIONCANCELLED:
		return "subscription-cancelled"
	case SUBSCRIPTIONFAILED:
		return "subscription-failed"
	default:
		return "invalid-status"
	}
}

func ParseInvoiceStatus(status string) (Status, error) {
	switch strings.ToLower(status) {
	case "new":
		return NEW, nil
	case "in-progress":
		return INPROGRESS, nil
	case "completed":
		return COMPLETED, nil
	case "failed":
		return FAILED, nil
	case "cancelled":
		return CANCELLED, nil
	case "subscription-active":
		return SUBSCRIPTIONACTIVE, nil
	case "subscription-expired":
		return SUBSCRIPTIONEXPIRED, nil
	case "subscription-cancelled":
		return SUBSCRIPTIONCANCELLED, nil
	case "subscription-failed":
		return SUBSCRIPTIONFAILED, nil
	default:
		return 0, errors.New("invalid invoice status")
	}
}
