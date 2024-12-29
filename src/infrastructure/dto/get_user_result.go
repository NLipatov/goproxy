package dto

import "goproxy/domain/aggregates"

type GetUserResult struct {
	Id       int
	Username string
}

func FromUser(user aggregates.User) GetUserResult {
	return GetUserResult{
		Id:       user.Id(),
		Username: user.Username(),
	}
}
