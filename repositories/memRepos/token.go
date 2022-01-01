package memRepos

import (
	Cerr "mmr/errors"
	"sync"
	"time"
)

type Token struct {
	storage map[string]int32
	mu      sync.Mutex
}

func NewToken(storage map[string]int32) *Token {
	return &Token{
		storage: storage,
		mu:      sync.Mutex{},
	}
}

func (t *Token) Get(uuid string) (int32, Cerr.CError) {
	t.mu.Lock()
	defer t.mu.Unlock()
	userID, ok := t.storage[uuid]
	if !ok {
		return -1, Cerr.NewUnauthorized("token")
	}

	return userID, nil
}

func (t *Token) Set(uuid string, userID int32, exp time.Duration) Cerr.CError {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.storage[uuid] = userID

	return nil
}

func (t *Token) Del(uuid string) Cerr.CError {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.storage, uuid)

	return nil
}
