package app

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v4"
	gcontext "mmr/context"
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

	id := gcontext.GetUserID(r.Context())
	row := conn.QueryRow(context.Background(),
		"SELECT id, name, email FROM users WHERE id = $1", id)

	var usr models.User
	if err = row.Scan(&usr.Id, &usr.Name, &usr.Email); err == pgx.ErrNoRows {
		http.Error(w, "", http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to SELECT: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err = json.NewEncoder(w).Encode(usr); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to encode json: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}
