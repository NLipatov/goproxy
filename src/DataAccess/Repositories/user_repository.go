package Repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"goproxy/Domain/Aggregates"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (u *UserRepository) GetByUsername(username string) (Aggregates.User, error) {
	var id int
	var usernameResult string
	var passwordHash, salt []byte

	err := u.db.
		QueryRow("SELECT id, username, password_hash, password_salt FROM public.users WHERE username = $1", username).
		Scan(&id, &usernameResult, &passwordHash, &salt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Aggregates.User{}, fmt.Errorf("user not found: %v", err)
		}
		return Aggregates.User{}, fmt.Errorf("could not load user: %v", err)
	}

	user, userErr := Aggregates.NewUser(id, username, passwordHash, salt)
	if userErr != nil {
		return Aggregates.User{}, fmt.Errorf("invalid user data: %v", userErr)
	}

	return user, nil
}

func (u *UserRepository) GetById(id int) (Aggregates.User, error) {
	var username string
	var passwordHash, salt []byte

	err := u.db.
		QueryRow("SELECT id, username, password_hash, password_salt FROM public.users WHERE id = $1", id).
		Scan(&id, &username, &passwordHash, &salt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Aggregates.User{}, fmt.Errorf("user not found: %v", err)
		}
		return Aggregates.User{}, fmt.Errorf("could not load user: %v", err)
	}

	user, userErr := Aggregates.NewUser(id, username, passwordHash, salt)
	if userErr != nil {
		return Aggregates.User{}, fmt.Errorf("invalid user data: %v", userErr)
	}

	return user, nil
}

func (u *UserRepository) Create(user Aggregates.User) (int, error) {
	var id int
	err := u.db.QueryRow("INSERT INTO public.users (username, password_hash, password_salt) VALUES ($1, $2, $3) RETURNING id",
		user.Username(), user.PasswordHash(), user.PasswordSalt(),
	).Scan(&id)
	return id, err
}

func (u *UserRepository) Update(user Aggregates.User) error {
	_, err := u.db.
		Exec("UPDATE public.users SET username = $1, password_hash = $2, password_salt = $3 WHERE id = $4",
			user.Username(), user.PasswordHash(), user.PasswordSalt(), user.Id())
	if err != nil {
		return fmt.Errorf("could not update user: %v", err)
	}
	return nil
}

func (u *UserRepository) Delete(user Aggregates.User) error {
	_, err := u.db.Exec("DELETE FROM public.users WHERE id = $1", user.Id())
	if err != nil {
		return fmt.Errorf("could not delete user: %v", err)
	}
	return nil
}
