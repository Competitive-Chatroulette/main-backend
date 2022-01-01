package redisRepos

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	Cerr "mmr/errors"
	"os"
	"strconv"
	"time"
)

type Token struct {
	rdb *redis.Client
}

func NewToken(rdb *redis.Client) *Token {
	return &Token{
		rdb: rdb,
	}
}

func (t *Token) Get(uuid string) (int32, Cerr.CError) {
	userIDStr, err := t.rdb.Get(context.TODO(), uuid).Result()
	if err == redis.Nil {
		return -1, Cerr.NewUnauthorized("token")
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't get token from redis: %v", err)
		return -1, Cerr.NewInternal()
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't convert redis str to int32: %v", err)
		return -1, Cerr.NewInternal()
	}

	return int32(userID), nil
}

func (t *Token) Set(uuid string, userID int32, exp time.Duration) Cerr.CError {
	err := t.rdb.Set(context.TODO(), uuid, strconv.Itoa(int(userID)), exp).Err()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't insert token into redis: %v", err)
		return Cerr.NewInternal()
	}

	return nil
}

func (t *Token) Del(uuid string) Cerr.CError {
	deleted, err := t.rdb.Del(context.TODO(), uuid).Result()
	if deleted == 0 {
		return Cerr.NewUnauthorized("token")
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't delete token from redis: %v", err)
		return Cerr.NewInternal()
	}

	return nil
}
