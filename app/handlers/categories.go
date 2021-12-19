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

	//skip rows with errors while scanning
	categories := make([]models.Category, 0)
	for rows.Next() {
		var category models.Category
		if err = rows.Scan(&category.Id, &category.Name); err == nil {
			categories = append(categories, category)
		}
	}

	//if none of the categories were scanned, return 500; else return what was successfully scanned
	if err := rows.Err(); err != nil {
		//log scanning errors
		fmt.Fprintf(os.Stderr, "Error while reading categories table: ", err)

		if len(categories) == 0 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
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
	if err != nil {
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

	var category models.Category
	err = row.Scan(&category.Id, &category.Name)
	if err == pgx.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to SELECT: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(category)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to encode json: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
