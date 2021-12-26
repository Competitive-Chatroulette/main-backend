package main

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"mmr/app"
	"mmr/services"
	postgresql "mmr/services/repositories"
	"os"
)

func main() {
	//init service
	p, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}
	defer p.Close()
	repo := postgresql.NewUser(p)
	svc := services.NewUser(repo)

	//init redis
	dsn := os.Getenv("REDIS_DSN")
	if len(dsn) == 0 {
		dsn = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: dsn,
	})
	_, err = rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal("Unable to connect to redis:", err)
	}

	a := app.NewApp(svc, rdb)
	a.Run()
}
