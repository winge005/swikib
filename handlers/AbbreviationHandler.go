package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"
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

func AbbreviationDelete(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}
	pathparam := strings.TrimPrefix(r.URL.Path, "/swiki/abbreviation/")
	persistence.DeleteAbbreviation(pathparam)

}

func AbbreviationAddHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	var errorMessage = ""
	var abbreviation model.Abbreviation
	r.ParseForm()

	abbreviation.Name = helpers.RemoveSpacesBeforAndAfter(r.FormValue("name"))
	abbreviation.Description = helpers.RemoveSpacesBeforAndAfter(r.FormValue("description"))

	if len(abbreviation.Name) < 2 || len(abbreviation.Description) < 2 {
		log.Println("Wrong input")
		errorMessage = "name or description not filled"
	}

	tmpl, err := template.ParseFiles("templates/addabbreviation.html")
	if tmpl == nil {
		errorMessage = err.Error()
	}

	if len(errorMessage) < 1 {
		abbrs, err := persistence.GetAbbreviationsForLetter(strings.ToUpper(abbreviation.Name[0:1]))
		if err != nil {
			errorMessage = err.Error()
		}

		for _, abbr := range abbrs {
			if strings.ToUpper(abbr.Name) == strings.ToUpper(abbreviation.Name) {
				errorMessage = "Name already used"
			}
		}
	}

	var recordNr int
	if len(errorMessage) < 1 {
		recordNr, err = persistence.AddAbbreviation(abbreviation)
		if err != nil {
			errorMessage = err.Error()
		}
	}

	data := struct {
		Abbreviation model.Abbreviation
		ErrorMessage string
		Message      string
	}{
		Abbreviation: abbreviation,
		ErrorMessage: errorMessage,
		Message:      "added abbreviation " + strconv.Itoa(recordNr),
	}

	err = tmpl.Execute(w, data)

	log.Println("Added: " + abbreviation.Name + " " + abbreviation.Description)
}
