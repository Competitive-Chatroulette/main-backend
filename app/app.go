package app

import (
	"github.com/gorilla/mux"
	"log"
	"mmr/services"
	"net/http"
)

type App struct {
	r       *mux.Router
	usrSvc  *services.User
	ctgSvc  *services.Category
	authSvc *services.Auth
}

func NewApp(usrSvc *services.User, ctgSvc *services.Category, authSvc *services.Auth) *App {
	a := &App{
		usrSvc:  usrSvc,
		ctgSvc:  ctgSvc,
		authSvc: authSvc,
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
	userR := a.r.PathPrefix("/users").Subrouter()
	userR.Use(a.withClaims)
	userR.HandleFunc("/me", a.getMe).Methods("GET")

	//CATEGORIES
	categR := a.r.PathPrefix("/categories").Subrouter()
	categR.HandleFunc("/", a.listCategories).Methods("GET")
	categR.HandleFunc("/{id:[0-9]+}", a.getCategory).Methods("GET")

	//AUTH
	authR := a.r.PathPrefix("/auth").Subrouter()
	authR.Use(a.withValidatedUser)
	authR.HandleFunc("/login", a.login).Methods("POST")
	authR.HandleFunc("/register", a.register).Methods("POST")

	tauthR := a.r.PathPrefix("/auth").Subrouter()
	tauthR.Use(a.withClaims)
	tauthR.HandleFunc("/logout", a.logout).Methods("POST")
	tauthR.HandleFunc("/refresh", a.refresh).Methods("POST")

	http.Handle("/", a.r)
}
