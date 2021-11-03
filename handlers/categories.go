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

type category struct {
	Id uint64
	Name string
}

func ListCategories(dbPool *pgxpool.Pool, w http.ResponseWriter, r *http.Request) {
	conn, err := dbPool.Acquire(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to acquire a database connection: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Release()

	rows, err := conn.Query(context.Background(), "SELECT id, name FROM categories")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to SELECT all: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	categories := make([]category, 0)
	for rows.Next() {
		var _category category
		if err = rows.Scan(&_category.Id, &_category.Name); err != nil {
			fmt.Fprintf(os.Stderr, "Unable to SELECT all: %v\n", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		categories = append(categories, _category)
	}

	if err := rows.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error while reading categories table: ", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(categories)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to encode json: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func GetCategory(dbPool *pgxpool.Pool, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil { // bad request
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	conn, err := dbPool.Acquire(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to acquire a database connection: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Release()

	row := conn.QueryRow(context.Background(),
		"SELECT id, name FROM categories WHERE id = $1", id)

	var _category category
	err = row.Scan(&_category.Id, &_category.Name)
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
	err = json.NewEncoder(w).Encode(_category)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to encode json: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
