package handlers

import (
	"log"
	"net/http"
	"swiki/helpers"
	"swiki/model"
	"swiki/persistence"
	"text/template"
)

func AbbreviationHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	log.Println("AbbreviationViewHandler")
	fl := r.PathValue("fl")

	var errorMessage = ""

	if len(fl) != 1 {
		errorMessage = "Wrong input"
	}

	abbreviations, err := persistence.GetAbbreviationsForLetter(fl)
	if err != nil {
		errorMessage = err.Error()
	}

	tmpl, err := template.ParseFiles("templates/abbreviations.html")
	if tmpl == nil {
		errorMessage = err.Error()
	}
	log.Println(errorMessage)
	data := struct {
		Abbreviations []model.Abbreviation
		ErrorMessage  string
	}{
		Abbreviations: abbreviations,
		ErrorMessage:  errorMessage,
	}
	err = tmpl.Execute(w, data)

	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func AbbreviationAddHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	log.Println("PageAbbreviationHandler")

	var abbreviation model.Abbreviation
	r.ParseForm()

	abbreviation.Name = r.FormValue("name")
	abbreviation.Description = r.FormValue("description")
	//title := r.FormValue("title")
	//content := r.FormValue("mdcontent")
	//tmpl, err := template.ParseFiles("templates/pageaddonsuccess.html")

	log.Println(abbreviation.Name + " " + abbreviation.Description)
}
