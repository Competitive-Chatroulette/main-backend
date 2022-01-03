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

//Set method operates on the assumption that it won't be called on the same uuid more than once
//Otherwise, a timer-based approach for deleting expired tokens needs to be implemented
func (t *Token) Set(uuid string, userID int32, exp time.Duration) Cerr.CError {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.storage[uuid] = userID
	//remove the token after expiration time has elapsed
	go func() {
		time.Sleep(exp)
		_ = t.Del(uuid)
	}()

	return nil
}

func (t *Token) Del(uuid string) Cerr.CError {
	t.mu.Lock()
	defer t.mu.Unlock()

	delete(t.storage, uuid)

	return nil
}
