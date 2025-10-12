package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"swiki/helpers"
	"swiki/model"
	"swiki/search"
)

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
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

	var query model.Search
	err = json.Unmarshal(bd, &query)

	if len(query.SearchQuery) == 0 {
		http.Error(w, "query is not filled", http.StatusInternalServerError)
		return
	}

	pagesFound, err := search.Search(query.SearchQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responseJson, err := json.Marshal(pagesFound)
	helpers.WriteResponse(w, string(responseJson))

}
