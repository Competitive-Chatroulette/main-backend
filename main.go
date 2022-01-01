package main

import (
	"mmr/app"
	"mmr/models"
	"mmr/repositories/memRepos"
	"mmr/services"
)

func main() {
	//pgx pool for pg repos
	/*p, err := pgxpool.Connect(context.TODO(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}*/
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
	usrRepo := memRepos.NewUser(make(map[int32]models.User), 0)
	usrSvc := services.NewUser(usrRepo)
	ctgRepo := memRepos.NewCategory(make(map[int32]models.Category))
	ctgSvc := services.NewCategory(ctgRepo)
	tokenRepo := memRepos.NewToken(make(map[string]int32))
	authSvc := services.NewAuth(usrRepo, tokenRepo)

	a := app.NewApp(usrSvc, ctgSvc, authSvc)
	a.Run()
}
