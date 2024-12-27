package DTOs

import "goproxy/Domain/Aggregates"

type GetUserDTO struct {
	Id       int
	Username string
}

func FromUser(user Aggregates.User) GetUserDTO {
	return GetUserDTO{
		Id:       user.Id(),
		Username: user.Username(),
	}
}
