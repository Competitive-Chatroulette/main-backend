package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"mmr/app/models"
	"net/http"
	"os"
)

func SignIn(p *pgxpool.Pool, w http.ResponseWriter, r *http.Request) {
	//TODO:validate required fields
	//serialize user
	var usr models.User
	err := json.NewDecoder(r.Body).Decode(&usr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid request: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//find the user in db
	conn, err := p.Acquire(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to acquire a database connection: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Release()
	row := conn.QueryRow(context.Background(),
		"SELECT id, name, email, pass FROM users WHERE email = $1", usr.Email)

	//serialize db user and return 404 if they don't exist
	var resUsr models.User
	err = row.Scan(&resUsr.Id, &resUsr.Name, &resUsr.Email, &resUsr.Pass)
	if err == pgx.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to SELECT: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//validate password
	err = resUsr.ValidatePass(usr.Pass)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Pass is not correct: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resUsr.Pass = "" //remove password hash from the user object we are returning

	//TODO: gen token and return it instead of the deserializing user
	//deserialize user
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(resUsr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to encode json: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func SignUp(p *pgxpool.Pool, w http.ResponseWriter, r *http.Request) {
	//TODO:validate required fields
	//serialize user
	var usr models.User
	err := json.NewDecoder(r.Body).Decode(&usr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid request: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//generate salted pass hash
	err = usr.HashPass(usr.Pass)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't hash the password: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//add user to db, get their id
	conn, err := p.Acquire(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to acquire a database connection: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Release()
	row := conn.QueryRow(context.Background(),
		"INSERT INTO users(name, email, pass) VALUES ($1, $2, $3) RETURNING id", usr.Name, usr.Email, usr.Pass)

	//return 400 if user is already in db
	//otherwise serialize new user id
	var id int
	err = row.Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			fmt.Fprintf(os.Stderr, pgErr.Message)
			w.WriteHeader(http.StatusConflict)
		} else {
			fmt.Fprintf(os.Stderr, "Unable to INSERT: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	//TODO: gen token and return it instead of the deserializing id
	//deserialize new user id
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to encode json: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
