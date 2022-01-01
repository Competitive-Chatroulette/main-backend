package memRepos

import (
	Cerr "mmr/errors"
	"mmr/models"
	"sync"
)

type User struct {
	storage   map[int32]models.User
	currentID int32
	mu        sync.Mutex
}

func NewUser(storage map[int32]models.User, startID int32) *User {
	return &User{
		storage:   storage,
		currentID: startID,
		mu:        sync.Mutex{},
	}
}

func (usr *User) Create(user *models.User) (int32, Cerr.CError) {
	usr.mu.Lock()
	defer usr.mu.Unlock()
	user.Id = usr.currentID
	usr.storage[usr.currentID] = *user
	usr.currentID += 1

	return usr.currentID - 1, nil
}

func (usr *User) FindById(userID int32) (*models.User, Cerr.CError) {
	usr.mu.Lock()
	defer usr.mu.Unlock()
	memUsr, ok := usr.storage[userID]
	if !ok {
		return nil, Cerr.NewNotFound("id")
	}
	return &memUsr, nil
}

func (usr *User) FindByEmail(email string) (*models.User, Cerr.CError) {
	usr.mu.Lock()
	defer usr.mu.Unlock()
	for _, memUsr := range usr.storage {
		if memUsr.Email == email {
			return &memUsr, nil
		}
	}

	return nil, Cerr.NewNotFound("email")
}
