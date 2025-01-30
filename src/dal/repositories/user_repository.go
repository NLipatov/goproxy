package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"goproxy/application"
	"goproxy/domain/aggregates"
)

type UserRepository struct {
	db    *sql.DB
	cache application.Cache[aggregates.User]
}

func NewUserRepository(db *sql.DB, cache application.Cache[aggregates.User]) *UserRepository {
	return &UserRepository{
		db:    db,
		cache: cache,
	}
}

func (u *UserRepository) GetByUsername(username string) (aggregates.User, error) {
	cachedUser, err := u.cache.Get(username)
	if err == nil {
		return cachedUser, nil
	}

	var id int
	var usernameResult string
	var emailResult string
	var passwordHash string

	err = u.db.
		QueryRow("SELECT id, username, email, password_hash FROM public.users WHERE username = $1", username).
		Scan(&id, &usernameResult, &emailResult, &passwordHash)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return aggregates.User{}, fmt.Errorf("user not found: %v", err)
		}
		return aggregates.User{}, fmt.Errorf("could not load user: %v", err)
	}

	user, userErr := aggregates.NewUser(id, usernameResult, emailResult, passwordHash)
	if userErr != nil {
		return aggregates.User{}, fmt.Errorf("invalid user %d stored in db: %v", id, userErr)
	}

	_ = u.cache.Set(username, user)

	return user, nil
}

func (u *UserRepository) GetById(id int) (aggregates.User, error) {
	cachedUser, cachedUserErr := u.cache.Get(fmt.Sprintf("%v", id))
	if cachedUserErr == nil {
		return cachedUser, nil
	}

	var username string
	var email string
	var passwordHash string

	err := u.db.
		QueryRow("SELECT id, username, email, password_hash FROM public.users WHERE id = $1", id).
		Scan(&id, &username, &email, &passwordHash)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return aggregates.User{}, fmt.Errorf("user not found: %v", err)
		}
		return aggregates.User{}, fmt.Errorf("could not load user: %v", err)
	}

	user, userErr := aggregates.NewUser(id, username, email, passwordHash)
	if userErr != nil {
		return aggregates.User{}, fmt.Errorf("invalid user data: %v", userErr)
	}

	_ = u.cache.Set(fmt.Sprintf("%v", id), user)

	return user, nil
}

func (u *UserRepository) GetByEmail(email string) (aggregates.User, error) {
	cachedUser, cachedUserErr := u.cache.Get(email)
	if cachedUserErr == nil {
		return cachedUser, nil
	}

	var id int
	var usernameResult string
	var emailResult string
	var passwordHash string

	err := u.db.
		QueryRow("SELECT id, username, email, password_hash FROM public.users WHERE email = $1", email).
		Scan(&id, &usernameResult, &emailResult, &passwordHash)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return aggregates.User{}, fmt.Errorf("user not found: %v", err)
		}
		return aggregates.User{}, fmt.Errorf("could not load user: %v", err)
	}

	user, userErr := aggregates.NewUser(id, usernameResult, emailResult, passwordHash)
	if userErr != nil {
		return aggregates.User{}, fmt.Errorf("invalid user data: %v", userErr)
	}

	_ = u.cache.Set(email, user)

	return user, nil
}

func (u *UserRepository) Create(user aggregates.User) (int, error) {
	var id int
	err := u.db.QueryRow("INSERT INTO public.users (username, email, password_hash) VALUES ($1, $2, $3) RETURNING id",
		user.Username(), user.Email(), user.PasswordHash(),
	).Scan(&id)
	return id, err
}

func (u *UserRepository) Update(user aggregates.User) error {
	result, err := u.db.
		Exec("UPDATE public.users SET username = $1, password_hash = $2 WHERE id = $3",
			user.Username(), user.PasswordHash(), user.Id())
	if err != nil {
		return fmt.Errorf("could not update user: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil || affected == 0 {
		return fmt.Errorf("no rows updated for user plan id: %d", user.Id())
	}

	_ = u.cache.Delete(fmt.Sprintf("%v", user.Id()))
	return nil
}

func (u *UserRepository) Delete(user aggregates.User) error {
	result, err := u.db.Exec("DELETE FROM public.users WHERE id = $1", user.Id())
	if err != nil {
		return fmt.Errorf("could not delete user: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil || affected == 0 {
		return fmt.Errorf("no rows affected")
	}
	return nil
}
