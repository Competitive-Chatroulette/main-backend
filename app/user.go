package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v4"
	"mmr/models"
	"net/http"
	"os"
)

func (a *App) GetUser(w http.ResponseWriter, r *http.Request) {
	conn, err := a.p.Acquire(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to acquire a database connection: %v\n", err)
		http.Error(w, "DB is busy", http.StatusInternalServerError)
		return
	}
	defer conn.Release()

	id := r.Context().Value("user_id")
	row := conn.QueryRow(context.Background(),
		"SELECT id, name, email FROM users WHERE id = $1", id)

	var usr models.User
	err = row.Scan(&usr.Id, &usr.Name, &usr.Email)
	if err == pgx.ErrNoRows {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
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
