package app

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"mmr/middleware"
	"net/http"
	"os"
)

type App struct {
	r *mux.Router
	p *pgxpool.Pool
}

func NewApp() *App {
	app := &App{}
	app.initDb()
	app.initRoutes()
	return app
}

func (a *App) Run() {
	defer a.p.Close()
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (a *App) initDb() {
	p, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}

	a.p = p
}

func (a *App) initRoutes() {
	a.r = mux.NewRouter()

	//USER
	userR := a.r.PathPrefix("/user").Subrouter()
	userR.Use(middleware.Authentication)
	userR.HandleFunc("/", a.GetUser).Methods("GET")

	//CATEGORIES
	categR := a.r.PathPrefix("/categories").Subrouter()
	categR.HandleFunc("/", a.ListCategories).Methods("GET")
	categR.HandleFunc("/{id:[0-9]+}", a.GetCategory).Methods("GET")

	//AUTH
	authR := a.r.PathPrefix("/auth").Subrouter()
	authR.HandleFunc("/signin", a.SignIn).Methods("POST")
	authR.HandleFunc("/signup", a.SignUp).Methods("POST")

	http.Handle("/", a.r)
}
