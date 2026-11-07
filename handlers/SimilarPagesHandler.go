package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"swiki/helpers"
	"swiki/model"
	"swiki/persistence"
)

func SimilarPagesHandler(w http.ResponseWriter, r *http.Request) {
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	bd, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("%s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var page model.Page
	if err := json.Unmarshal(bd, &page); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if len(page.Title) == 0 && len(page.Content) == 0 {
		http.Error(w, "title or content is required", http.StatusBadRequest)
		return
	}

	similar, err := persistence.FindSimilarPages(page, 10)
	if err != nil {
		log.Printf("could not find similar pages: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responseJson, err := json.Marshal(similar)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	helpers.WriteResponse(w, string(responseJson))
}

func PagesFTSStatusHandler(w http.ResponseWriter, r *http.Request) {
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	status := persistence.GetPagesFTSStatus()
	responseJson, err := json.Marshal(status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	helpers.WriteResponse(w, string(responseJson))
}
