package main

import (
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/net/context"
	"log"
	"mmr/app"
	"mmr/repositories/memRepos"
	"mmr/repositories/pgRepos"
	"mmr/services"
	"os"
)

func main() {
	//pgx pool for pg repos
	p, err := pgxpool.Connect(context.TODO(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}
	//redis for redis repos TODO: Test redis repo before uncommenting this
	/*dsn := os.Getenv("REDIS_DSN")
	if len(dsn) == 0 {
		dsn = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: dsn,
	})
	_, err = rdb.Ping(context.TODO()).Result()
	if err != nil {
		log.Fatal("Unable to connect to redis:", err)
	}*/
	//init services
	usrRepo := pgRepos.NewUser(p)
	usrSvc := services.NewUser(usrRepo)
	ctgRepo := pgRepos.NewCategory(p)
	ctgSvc := services.NewCategory(ctgRepo)
	tokenRepo := memRepos.NewToken()
	authSvc := services.NewAuth(usrRepo, tokenRepo)

	a := app.NewApp(usrSvc, ctgSvc, authSvc)
	a.Run()
}
