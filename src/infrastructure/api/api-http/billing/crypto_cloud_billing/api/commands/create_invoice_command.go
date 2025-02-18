package commands

import (
	"goproxy/domain/valueobjects"
)

type CreateInvoiceCommand struct {
	email    valueobjects.Email
	planId   int
	currency string
}

func NewCreateInvoiceCommand(email valueobjects.Email,
	planId int,
	currency string) CreateInvoiceCommand {
	return CreateInvoiceCommand{
		email:    email,
		planId:   planId,
		currency: currency,
	}
}

func (c *CreateInvoiceCommand) Email() valueobjects.Email {
	return c.email
}

func (c *CreateInvoiceCommand) PlanId() int {
	return c.planId
}

func (c *CreateInvoiceCommand) Currency() string {
	return c.currency
}
