package use_cases

import (
	"fmt"
	"goproxy/application/commands"
	"goproxy/application/contracts"
	"goproxy/domain/aggregates"
	"strconv"
	"strings"
)

type UserUseCasesContract interface {
	GetByEmail(email string) (aggregates.User, error)
	GetById(id int) (aggregates.User, error)
	Create(command commands.PostUser) (int, error)
	Update(entity aggregates.User) error
	Delete(dto commands.DeleteUser) error
}

type UserUseCases struct {
	repo          contracts.UserRepository
	cryptoService contracts.CryptoService
}

func NewUserUseCases(repo contracts.UserRepository, cryptoService contracts.CryptoService) UserUseCases {
	return UserUseCases{
		repo:          repo,
		cryptoService: cryptoService,
	}
}

func (u UserUseCases) GetById(id int) (aggregates.User, error) {
	return u.repo.GetById(id)
}

func (u UserUseCases) GetByEmail(email string) (aggregates.User, error) {
	return u.repo.GetByEmail(email)
}

func (u UserUseCases) Create(command commands.PostUser) (int, error) {
	hash, err := u.cryptoService.HashValue(command.Password.Value)
	if err != nil {
		return 0, err
	}

	user, err := aggregates.NewUser(-1, command.Username, command.Email, hash)
	if err != nil {
		return 0, err
	}
	return u.repo.Create(user)
}

func (u UserUseCases) Update(entity aggregates.User) error {
	return u.repo.Update(entity)
}

func (u UserUseCases) Delete(dto commands.DeleteUser) error {
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

	isPasswordValid := u.cryptoService.ValidateHash(user.PasswordHash(), dto.Password.Value)
	if !isPasswordValid {
		return fmt.Errorf("invalid password")
	}

	return u.repo.Delete(user)
}
