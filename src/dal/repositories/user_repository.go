package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"goproxy/domain/aggregates"
)

type UserRepository struct {
	db    *sql.DB
	cache BigCacheUserRepositoryCache
}

func NewUserRepository(db *sql.DB, cache BigCacheUserRepositoryCache) *UserRepository {
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
	var passwordHash, salt []byte

	err = u.db.
		QueryRow("SELECT id, username, password_hash, password_salt FROM public.users WHERE username = $1", username).
		Scan(&id, &usernameResult, &passwordHash, &salt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return aggregates.User{}, fmt.Errorf("user not found: %v", err)
		}
		return aggregates.User{}, fmt.Errorf("could not load user: %v", err)
	}

	user, userErr := aggregates.NewUser(id, username, passwordHash, salt)
	if userErr != nil {
		return aggregates.User{}, fmt.Errorf("invalid user data: %v", userErr)
	}

	_ = u.cache.Set(username, user)

	return user, nil
}

func (u *UserRepository) GetById(id int) (aggregates.User, error) {
	var username string
	var passwordHash, salt []byte

	err := u.db.
		QueryRow("SELECT id, username, password_hash, password_salt FROM public.users WHERE id = $1", id).
		Scan(&id, &username, &passwordHash, &salt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return aggregates.User{}, fmt.Errorf("user not found: %v", err)
		}
		return aggregates.User{}, fmt.Errorf("could not load user: %v", err)
	}

	user, userErr := aggregates.NewUser(id, username, passwordHash, salt)
	if userErr != nil {
		return aggregates.User{}, fmt.Errorf("invalid user data: %v", userErr)
	}

	return user, nil
}

func (u *UserRepository) Create(user aggregates.User) (int, error) {
	var id int
	err := u.db.QueryRow("INSERT INTO public.users (username, password_hash, password_salt) VALUES ($1, $2, $3) RETURNING id",
		user.Username(), user.PasswordHash(), user.PasswordSalt(),
	).Scan(&id)
	return id, err
}

func (u *UserRepository) Update(user aggregates.User) error {
	result, err := u.db.
		Exec("UPDATE public.users SET username = $1, password_hash = $2, password_salt = $3 WHERE id = $4",
			user.Username(), user.PasswordHash(), user.PasswordSalt(), user.Id())
	if err != nil {
		return fmt.Errorf("could not update user: %v", err)
	}

	affected, err := result.RowsAffected()
	if err != nil || affected == 0 {
		return fmt.Errorf("no rows updated for user plan id: %d", user.Id())
	}

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
