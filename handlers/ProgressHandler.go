package handlers

import (
	"encoding/json"
	"net/http"
	"swiki/persistencelocal"
)

func ProgressHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(persistencelocal.CurrentProgress()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
