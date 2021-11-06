package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"mmr/app/models"
	"net/http"
	"os"
	"strconv"
)

func ListUsers(dbPool *pgxpool.Pool, w http.ResponseWriter, r *http.Request) {
	conn, err := dbPool.Acquire(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to acquire a database connection: %v\n", err)
		http.Error(w, "DB is busy", http.StatusInternalServerError)
		return
	}
	defer conn.Release()

	rows, err := conn.Query(context.Background(), "SELECT id, name, email FROM users")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to SELECT all: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	users := make([]models.User, 0)
	for rows.Next() {
		var usr models.User
		err = rows.Scan(&usr.Id, &usr.Name, &usr.Email)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to SELECT all: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		users = append(users, usr)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(users)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to encode json: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func GetUser(dbPool *pgxpool.Pool, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil { // bad request
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	conn, err := dbPool.Acquire(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to acquire a database connection: %v\n", err)
		http.Error(w, "DB is busy", http.StatusInternalServerError)
		return
	}
	defer conn.Release()

	row := conn.QueryRow(context.Background(),
		"SELECT id, name, email FROM users WHERE id = $1", id)

	var usr models.User
	err = row.Scan(&usr.Id, &usr.Name, &usr.Email)
	if err == pgx.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to SELECT: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(usr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to encode json: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
