package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"net/http"
	"os"
	"strconv"
)

type user struct {
	Id uint64 `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
}

func ListUsers(dbPool *pgxpool.Pool, w http.ResponseWriter, r *http.Request) {
	conn, err := dbPool.Acquire(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to acquire a database connection: %v\n", err)
		http.Error(w, "DB is busy", 500)
		return
	}
	defer conn.Release()

	rows, err := conn.Query(context.Background(), "SELECT id, name, email FROM users")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to SELECT all: %v\n", err)
		w.WriteHeader(500)
		return
	}

	users := make([]user, 0)
	for rows.Next() {
		var usr user
		err = rows.Scan(&usr.Id, &usr.Name, &usr.Email)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to SELECT all: %v\n", err)
			w.WriteHeader(500)
			return
		}

		users = append(users, usr)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(users)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to encode json: %v", err)
		w.WriteHeader(500)
		return
	}
}

func GetUser(dbPool *pgxpool.Pool, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil { // bad request
		http.Error(w, err.Error(), 400)
		return
	}

	conn, err := dbPool.Acquire(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to acquire a database connection: %v\n", err)
		http.Error(w, "DB is busy", 500)
		return
	}
	defer conn.Release()

	row := conn.QueryRow(context.Background(),
		"SELECT id, name, email FROM users WHERE id = $1", id)

	var usr user
	err = row.Scan(&usr.Id, &usr.Name, &usr.Email)
	if err == pgx.ErrNoRows {
		w.WriteHeader(404)
		return
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to SELECT: %v", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(usr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to encode json: %v", err)
		w.WriteHeader(500)
		return
	}
}
