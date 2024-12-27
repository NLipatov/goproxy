package DTO

import "goproxy/Domain/Aggregates"

type GetUserQuery struct {
	Id       int
	Username string
}

func FromUser(user Aggregates.User) GetUserQuery {
	return GetUserQuery{
		Id:       user.Id(),
		Username: user.Username(),
	}
}
