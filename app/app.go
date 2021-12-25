package app

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"net/http"
	"os"
)

type App struct {
	r   *mux.Router
	p   *pgxpool.Pool
	rdb *redis.Client
}

func NewApp() *App {
	a := &App{}

	//init pool
	p, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}
	a.p = p

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
	defer a.p.Close()
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
