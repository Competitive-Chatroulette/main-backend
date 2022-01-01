package models

import (
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id    int32  `json:"id"` //TODO: UserID should be changed to uint64 everywhere
	Name  string `json:"name,omitempty" validate:"lte=20"`
	Email string `json:"email,omitempty" validate:"required,email"`
	Pass  string `json:"pass,omitempty" validate:"required,gte=6"`
}

func (usr *User) HashPass(pass string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	usr.Pass = string(hash)
	return nil
}

func (usr *User) ValidatePass(pass string) error {
	return bcrypt.CompareHashAndPassword([]byte(usr.Pass), []byte(pass))
}
