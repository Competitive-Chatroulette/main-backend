package app

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"mmr/services"
	postgresql "mmr/services/repositories"
	"net/http"
	"os"
)

type App struct {
	r      *mux.Router
	usrSvc *services.User
	ctgSvc *services.Category
	rdb    *redis.Client
}

func NewApp() *App {
	a := &App{}
	//init services
	p, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}
	usrRepo := postgresql.NewUser(p)
	a.usrSvc = services.NewUser(usrRepo)
	ctgRepo := postgresql.NewCategory(p)
	a.ctgSvc = services.NewCategory(ctgRepo)

	//init redis
	dsn := os.Getenv("REDIS_DSN")
	if len(dsn) == 0 {
		dsn = "localhost:6379"
	}
	a.rdb = redis.NewClient(&redis.Options{
		Addr: dsn,
	})
	_, err = a.rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal("Unable to connect to redis:", err)
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
	categR := a.r.PathPrefix("/categories").Subrouter()
	categR.HandleFunc("/", a.ListCategories).Methods("GET")
	categR.HandleFunc("/{id:[0-9]+}", a.GetCategory).Methods("GET")

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
