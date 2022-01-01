package main

import (
	"mmr/app"
	"mmr/models"
	"mmr/repositories/memRepos"
	"mmr/services"
)

func main() {
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
