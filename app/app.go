package app

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	"log"
	"mmr/handlers"
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

func (app *App) Run() {
	defer app.p.Close()
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (app *App) initDb() {
	p, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}

	app.p = p
}

func (app *App) initRoutes() {
	app.r = mux.NewRouter()

	//USERS
	usersR := app.r.PathPrefix("/users").Subrouter()
	usersR.HandleFunc("/", withPool(app.p, handlers.ListUsers)).Methods("GET")
	usersR.HandleFunc("/{id:[0-9]+}", withPool(app.p, handlers.GetUser)).Methods("GET")

	//CATEGORIES
	categoriesR := app.r.PathPrefix("/categories").Subrouter()
	categoriesR.HandleFunc("/", withPool(app.p, handlers.ListCategories)).Methods("GET")
	categoriesR.HandleFunc("/{id:[0-9]+}", withPool(app.p, handlers.GetCategory)).Methods("GET")

	//AUTH
	authR := app.r.PathPrefix("/auth").Subrouter()
	authR.HandleFunc("/signin", withPool(app.p, handlers.SignIn)).Methods("POST")
	authR.HandleFunc("/signup", withPool(app.p, handlers.SignUp)).Methods("POST")

	http.Handle("/", app.r)
}

func withPool(p *pgxpool.Pool, handler func(*pgxpool.Pool, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handler(p, w, r)
	}
}

