package repositories

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"goproxy/application"
	"goproxy/domain"
	"goproxy/domain/aggregates"
	"goproxy/domain/events"
	"goproxy/infrastructure/config"
	"log"
	"strings"
	"sync"
)

type UserRepository struct {
	db          *sql.DB
	cache       BigCacheUserRepositoryCache
	kafkaConfig config.KafkaConfig
	messageBus  application.MessageBusService
	once        sync.Once
}

func NewUserRepository(db *sql.DB, cache BigCacheUserRepositoryCache, messageBusService application.MessageBusService) *UserRepository {
	service := &UserRepository{
		db:         db,
		cache:      cache,
		messageBus: messageBusService,
	}

	service.startProcessingEvents()

	return service
}

func (u *UserRepository) startProcessingEvents() {
	u.once.Do(func() {

		go u.processEvents()
	})
}

func (u *UserRepository) processEvents() {
	defer func(messageBus application.MessageBusService) {
		_ = messageBus.Close()
	}(u.messageBus)

	topics := []string{fmt.Sprintf("%s", domain.AUTH)}
	err := u.messageBus.Subscribe(topics)
	if err != nil {
		log.Fatalf("Failed to subscribe to topics: %s", err)
	}

	log.Printf("Subscribed to topics: %s", strings.Join(topics, ", "))

	for {
		event, consumeErr := u.messageBus.Consume()
		if consumeErr != nil {
			log.Printf("failed to consume from message bus: %s", consumeErr)
		}

		if event == nil {
			log.Printf("received nil event from message bus")
			continue
		}

		if event.EventType.Value() == "UserPasswordChangedEvent" {
			var userPasswordChangedEvent events.UserPasswordChangedEvent
			deserializationErr := json.Unmarshal([]byte(event.Payload), &userPasswordChangedEvent)
			if deserializationErr != nil {
				log.Printf("failed to deserialize user password changed event: %s", deserializationErr)
			}

			err = u.cache.Delete(userPasswordChangedEvent.Username)
			if err == nil {
				log.Printf("user %s removed from user repository cache", userPasswordChangedEvent.Username)
			} else {
				log.Printf("failed to delete user password changed event: %s", err)
			}
		}
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

	return user, nil
}

func (u *UserRepository) GetByEmail(email string) (aggregates.User, error) {
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
