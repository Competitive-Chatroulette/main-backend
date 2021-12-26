package app

import (
	"encoding/json"
	"fmt"
	gcontext "mmr/context"
	"net/http"
	"os"
)

func (a *App) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := gcontext.GetUserID(r.Context())
	dbUsr, cerr := a.usrSvc.FindById(userID)
	if cerr != nil {
		http.Error(w, cerr.Error(), cerr.GetStatusCode())
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(dbUsr); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to encode json: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}
