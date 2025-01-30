package cache_serialization

import (
	"goproxy/domain/aggregates"
)

type StoredUser struct {
	Id           int    `msgpack:"Id"`
	Username     string `msgpack:"Username"`
	Email        string `msgpack:"Email"`
	PasswordHash string `msgpack:"PasswordHash"`
}

type UserSerializer struct {
}

func NewUserSerializer() CacheSerializer[aggregates.User, StoredUser] {
	return UserSerializer{}
}

func (u UserSerializer) ToT(dto StoredUser) aggregates.User {
	user, _ := aggregates.NewUser(dto.Id, dto.Username, dto.Email, dto.PasswordHash)
	return user
}

func (u UserSerializer) ToD(user aggregates.User) StoredUser {
	return StoredUser{
		Id:           user.Id(),
		Username:     user.Username(),
		Email:        user.Email(),
		PasswordHash: user.PasswordHash(),
	}
}

func (u UserSerializer) ToTArray(dto []StoredUser) []aggregates.User {
	result := make([]aggregates.User, len(dto))
	for i, v := range dto {
		result[i] = u.ToT(v)
	}

	return result
}

func (u UserSerializer) ToDArray(users []aggregates.User) []StoredUser {
	result := make([]StoredUser, len(users))
	for i, v := range users {
		result[i] = u.ToD(v)
	}

	return result
}
