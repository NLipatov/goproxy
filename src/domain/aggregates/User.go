package aggregates

import "goproxy/domain/valueobjects"

type User struct {
	id           int
	username     valueobjects.Username
	email        valueobjects.Email
	passwordHash valueobjects.Hash
	passwordSalt valueobjects.Salt
}

func NewUser(id int, username string, email string, hash, salt []byte) (User, error) {
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

	saltObject, saltErr := valueobjects.NewSalt(salt)
	if saltErr != nil {
		return User{}, saltErr
	}

	return User{
		id:           id,
		username:     usernameObject,
		email:        emailObject,
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

func (u *User) Email() string {
	return u.email.String()
}

func (u *User) PasswordSalt() []byte {
	return u.passwordSalt.Value
}

func (u *User) PasswordHash() []byte {
	return u.passwordHash.Value
}
