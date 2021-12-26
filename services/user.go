package services

import (
	cerr "mmr/errors"
	"mmr/models"
)

type UserRepository interface {
	Create(user *models.User) (int32, cerr.CError)
	FindById(userID int32) (*models.User, cerr.CError)
	FindByEmail(email string) (*models.User, cerr.CError)
}

type User struct {
	repo UserRepository
}

func NewUser(repo UserRepository) *User {
	return &User{
		repo: repo,
	}
}

func (usr *User) Create(user *models.User) (int32, cerr.CError) {
	return usr.repo.Create(user)
}

func (usr *User) FindById(userID int32) (*models.User, cerr.CError) {
	return usr.repo.FindById(userID)
}

func (usr *User) FindByEmail(email string) (*models.User, cerr.CError) {
	return usr.repo.FindByEmail(email)
}
