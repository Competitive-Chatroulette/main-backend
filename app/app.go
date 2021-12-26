package app

import (
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"log"
	"mmr/services"
	"net/http"
)

type App struct {
	r      *mux.Router
	usrSvc *services.User
	rdb    *redis.Client
}

func NewApp(usrSrv *services.User, rdb *redis.Client) *App {
	a := &App{
		usrSvc: usrSrv,
		rdb:    rdb,
	}

	a.initRoutes()
	return a
}

func (a *App) Run() {
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (a *App) initRoutes() {
	a.r = mux.NewRouter()

	//USER
	userR := a.r.PathPrefix("/user").Subrouter()
	userR.Use(a.withToken)
	userR.Use(a.withClaims)
	userR.HandleFunc("/", a.GetUser).Methods("GET")

	//CATEGORIES
	/*	categR := a.r.PathPrefix("/categories").Subrouter()
		categR.HandleFunc("/", a.ListCategories).Methods("GET")
		categR.HandleFunc("/{id:[0-9]+}", a.GetCategory).Methods("GET")*/

	//AUTH
	authR := a.r.PathPrefix("/auth").Subrouter()
	authR.HandleFunc("/login", a.login).Methods("POST")
	authR.HandleFunc("/register", a.register).Methods("POST")
	tauthR := a.r.PathPrefix("/auth").Subrouter()
	tauthR.Use(a.withToken)
	tauthR.Use(a.withClaims)
	tauthR.HandleFunc("/logout", a.logout).Methods("POST")
	tauthR.HandleFunc("/refresh", a.refresh).Methods("POST")

	http.Handle("/", a.r)
}
