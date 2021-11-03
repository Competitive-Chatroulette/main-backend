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
	"net/http"
	"os"
)

type authUser struct {
	Name string `json:"name"`
	Email string `json:"email"`
	Pass string `json:"pass"`
}

func SignIn(p *pgxpool.Pool, w http.ResponseWriter, r *http.Request) {
	var usr authUser
	err := json.NewDecoder(r.Body).Decode(&usr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid request: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	conn, err := p.Acquire(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to acquire a database connection: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Release()

	row := conn.QueryRow(context.Background(),
		"SELECT id, name, email FROM users WHERE email = $1 AND pass = $2", usr.Email, usr.Pass)

	fmt.Println(row)
	var resUsr user
	err = row.Scan(&resUsr.Id, &resUsr.Name, &resUsr.Email)
	if err == pgx.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to SELECT: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(resUsr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to encode json: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func SignUp(p *pgxpool.Pool, w http.ResponseWriter, r *http.Request) {
	var usr authUser
	err := json.NewDecoder(r.Body).Decode(&usr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid request: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	conn, err := p.Acquire(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to acquire a database connection: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Release()

	row := conn.QueryRow(context.Background(),
		"INSERT INTO users(name, email, pass) VALUES ($1, $2, $3) RETURNING id", usr.Name, usr.Email, usr.Pass)

	var id int
	err = row.Scan(&id)
	//this is embarrassing, I just wanted to check whether the error is unique constraint violation or not
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			fmt.Fprintf(os.Stderr, pgErr.Message)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		fmt.Fprintf(os.Stderr, "Unable to INSERT: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to encode json: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
