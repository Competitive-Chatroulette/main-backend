package app

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"strconv"
)

func (a *App) listCategories(w http.ResponseWriter, r *http.Request) {
	ctgs, cerr := a.ctgSvc.List()
	if cerr != nil {
		http.Error(w, cerr.Error(), cerr.GetStatusCode())
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(ctgs); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to encode json: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
	}
}

func (a *App) getCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 32)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctg, cerr := a.ctgSvc.Get(int32(id))
	if cerr != nil {
		http.Error(w, cerr.Error(), cerr.GetStatusCode())
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err = json.NewEncoder(w).Encode(ctg); err != nil {
		fmt.Fprintf(os.Stderr, "Unable to encode json: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
	}
}
