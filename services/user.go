package services

import (
	Cerr "mmr/errors"
	"mmr/models"
)

type UserRepository interface {
	Create(user *models.User) (int32, Cerr.CError)
	FindById(userID int32) (*models.User, Cerr.CError)
	FindByEmail(email string) (*models.User, Cerr.CError)
}

type User struct {
	repo UserRepository
}

func NewUser(repo UserRepository) *User {
	return &User{
		repo: repo,
	}
}

func (usr *User) Create(user *models.User) (int32, Cerr.CError) {
	return usr.repo.Create(user)
}

func (usr *User) Find(userID int32) (*models.User, Cerr.CError) {
	dbUsr, cerr := usr.repo.FindById(userID)
	if cerr != nil {
		return nil, cerr
	}

	//remove password from the model
	dbUsr.Pass = ""

	return dbUsr, nil
}
