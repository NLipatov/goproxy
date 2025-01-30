package cache_serialization

import (
	"goproxy/domain/aggregates"
)

type UserDto struct {
	Id           int
	Username     string
	Email        string
	PasswordHash string
}

type UserSerializer struct {
}

func NewUserSerializer() CacheSerializer[aggregates.User, UserDto] {
	return UserSerializer{}
}

func (u UserSerializer) ToT(dto UserDto) aggregates.User {
	user, _ := aggregates.NewUser(dto.Id, dto.Username, dto.Email, dto.PasswordHash)
	return user
}

func (u UserSerializer) ToD(user aggregates.User) UserDto {
	return UserDto{
		Id:           user.Id(),
		Username:     user.Username(),
		Email:        user.Email(),
		PasswordHash: user.PasswordHash(),
	}
}

func (u UserSerializer) ToTArray(dto []UserDto) []aggregates.User {
	result := make([]aggregates.User, len(dto))
	for i, v := range dto {
		result[i] = u.ToT(v)
	}

	return result
}

func (u UserSerializer) ToDArray(users []aggregates.User) []UserDto {
	result := make([]UserDto, len(users))
	for i, v := range users {
		result[i] = u.ToD(v)
	}

	return result
}
