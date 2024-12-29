package aggregates

import "goproxy/domain/valueobjects"

type User struct {
	id           int
	username     valueobjects.Username
	passwordHash valueobjects.Hash
	passwordSalt valueobjects.Salt
}

func NewUser(id int, username string, hash, salt []byte) (User, error) {
	usernameObject, usernameErr := valueobjects.NewUsernameFromString(username)
	if usernameErr != nil {
		return User{}, usernameErr
	}

	passwordHashObject, hashErr := valueobjects.NewHash(hash)
	if hashErr != nil {
		return User{}, hashErr
	}

	saltObject, saltErr := valueobjects.NewSalt(salt)
	if saltErr != nil {
		return User{}, saltErr
	}

	return User{
		id:           id,
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
