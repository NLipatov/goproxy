package Application

import "goproxy/Domain/Aggregates"

type UserUseCases struct {
	repo Repository[Aggregates.User]
}

func NewUserUseCases(repo Repository[Aggregates.User]) *UserUseCases {
	return &UserUseCases{
		repo: repo,
	}
}

func (u UserUseCases) GetById(id int) (Aggregates.User, error) {
	return u.repo.GetById(id)
}

func (u UserUseCases) Create(entity Aggregates.User) (int, error) {
	return u.repo.Create(entity)
}

func (u UserUseCases) Update(entity Aggregates.User) error {
	return u.repo.Update(entity)
}

func (u UserUseCases) Delete(entity Aggregates.User) error {
	return u.repo.Delete(entity)
}
