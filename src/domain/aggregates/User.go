package aggregates

import "goproxy/domain/valueobjects"

type User struct {
	id           int
	username     valueobjects.Username
	email        valueobjects.Email
	passwordHash valueobjects.Argon2idHash
}

func NewUser(id int, username, email, hash string) (User, error) {
	usernameObject, usernameObjectErr := valueobjects.NewUsernameFromString(username)
	if usernameObjectErr != nil {
		return User{}, usernameObjectErr
	}

	emailObject, emailObjectErr := valueobjects.ParseEmailFromString(email)
	if emailObjectErr != nil {
		return User{}, emailObjectErr
	}

	passwordHashObject, hashErr := valueobjects.NewHash(hash)
	if hashErr != nil {
		return User{}, hashErr
	}

	return User{
		id:           id,
		username:     usernameObject,
		email:        emailObject,
		passwordHash: passwordHashObject,
	}, nil
}

func (u *User) Id() int {
	return u.id
}

func (u *User) Username() string {
	return u.username.Value
}

func (u *User) Email() string {
	return u.email.String()
}

func (u *User) PasswordHash() string {
	return u.passwordHash.Value
}
