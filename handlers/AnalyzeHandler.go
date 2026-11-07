package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"swiki/persistencelocal"
)

func AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	if !persistencelocal.TryStartProgress("analyse") {
		http.Error(w, "Er loopt al een proces", http.StatusConflict)
		return
	}

	defer func() {
		if rec := recover(); rec != nil {
			persistencelocal.FinishProgress(fmt.Errorf("%v", rec))
			panic(rec)
		}
	}()

	pairs, err := persistencelocal.FindSimilarPagePairs()
	if err != nil {
		persistencelocal.FinishProgress(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(pairs); err != nil {
		persistencelocal.FinishProgress(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	persistencelocal.FinishProgress(nil)
}
