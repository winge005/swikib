package handlers

import (
	"log"
	"net/http"
	"swiki/model"
	"swiki/search"
	"text/template"
)

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("SearchHandler")

	var errorMessage = ""

	r.ParseForm()
	query := r.FormValue("query")
	log.Println(query)

	pages, err := search.Search(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/search.html")
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(pages) == 0 {
		errorMessage = "No results found."
	}

	data := struct {
		Query        string
		ErrorMessage string
		Pages        []model.Page
	}{
		Query:        query,
		ErrorMessage: errorMessage,
		Pages:        pages,
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
