package Application

import (
	"fmt"
	"goproxy/Application/Commands"
	"goproxy/Domain/Aggregates"
	"strconv"
	"strings"
)

type UserUseCases struct {
	repo          Repository[Aggregates.User]
	cryptoService CryptoService
}

func NewUserUseCases(repo Repository[Aggregates.User], cryptoService CryptoService) UserUseCases {
	return UserUseCases{
		repo:          repo,
		cryptoService: cryptoService,
	}
}

func (u UserUseCases) GetById(id int) (Aggregates.User, error) {
	return u.repo.GetById(id)
}

func (u UserUseCases) Create(command Commands.PostUserCommand) (int, error) {
	salt, err := u.cryptoService.GenerateSalt()
	if err != nil {
		return 0, err
	}

	hash, err := u.cryptoService.HashValue(command.Password.Value, salt)
	if err != nil {
		return 0, err
	}

	user, err := Aggregates.NewUser(-1, command.Username, hash, salt)
	if err != nil {
		return 0, err
	}
	return u.repo.Create(user)
}

func (u UserUseCases) Update(entity Aggregates.User) error {
	return u.repo.Update(entity)
}

func (u UserUseCases) Delete(dto Commands.DeleteUserCommand) error {
	id, err := strconv.Atoi(dto.Id)
	if err != nil {
		return fmt.Errorf("invalid user id: %s", dto.Id)
	}
	user, err := u.repo.GetById(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("user not found")
		}
		return err
	}

	isPasswordValid := u.cryptoService.ValidateHash(user.PasswordHash(), user.PasswordSalt(), dto.Password.Value)
	if !isPasswordValid {
		return fmt.Errorf("invalid password")
	}

	return u.repo.Delete(user)
}
