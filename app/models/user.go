package models

import "golang.org/x/crypto/bcrypt"

type User struct {
	Id    uint64 `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	Pass  string `json:"pass,omitempty"`
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
