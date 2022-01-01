package memRepos

import (
	Cerr "mmr/errors"
	"time"
)

type Token struct {
	storage map[string]int32
}

func NewToken() *Token {
	return &Token{
		storage: make(map[string]int32),
	}
}

func (t *Token) Get(uuid string) (int32, Cerr.CError) {
	userID, ok := t.storage[uuid]
	if !ok {
		return -1, Cerr.NewUnauthorized("token")
	}

	return userID, nil
}

func (t *Token) Set(uuid string, userID int32, exp time.Duration) Cerr.CError {
	t.storage[uuid] = userID
	return nil
}

func (t *Token) Del(uuid string) Cerr.CError {
	delete(t.storage, uuid)
	return nil
}
