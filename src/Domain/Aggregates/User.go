package Aggregates

import "goproxy/Domain/ValueObjects"

type User struct {
	id           int
	username     ValueObjects.Username
	passwordHash ValueObjects.Hash
	passwordSalt ValueObjects.Salt
}

func NewUser(username string, hash, salt []byte) (User, error) {
	usernameObject, usernameErr := ValueObjects.NewUsernameFromString(username)
	if usernameErr != nil {
		return User{}, usernameErr
	}

	passwordHashObject, hashErr := ValueObjects.NewHash(hash)
	if hashErr != nil {
		return User{}, hashErr
	}

	saltObject, saltErr := ValueObjects.NewSalt(salt)
	if saltErr != nil {
		return User{}, saltErr
	}

	return User{
		id:           -1,
		username:     usernameObject,
		passwordHash: passwordHashObject,
		passwordSalt: saltObject,
	}, nil
}

func (u *User) Id() int {
	return u.id
}

func (u *User) Username() string {
	return u.username.Value
}

func (u *User) PasswordSalt() []byte {
	return u.passwordSalt.Value
}

func (u *User) PasswordHash() []byte {
	return u.passwordHash.Value
}
