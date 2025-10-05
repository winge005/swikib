package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"swiki/helpers"
	"swiki/model"
	"swiki/persistence"
)

func AbbreviationHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	log.Println("AbbreviationViewHandler")
	firstLetter := r.PathValue("fl")

	if len(firstLetter) != 1 {
		http.Error(w, "firstletter is not ok", http.StatusBadRequest)
		return
	}

	abbreviations, err := persistence.GetAbbreviationsForLetter(firstLetter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	response, err := json.Marshal(abbreviations)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	helpers.WriteResponse(w, string(response))
}

func AbbreviationDelete(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}
	pathparam := strings.TrimPrefix(r.URL.Path, "/swiki/abbreviation/")
	persistence.DeleteAbbreviation(pathparam)

	helpers.WriteResponse(w, "")
}

func AbbreviationAddHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	bd, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var abbreviation model.Abbreviation
	err = json.Unmarshal(bd, &abbreviation)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(abbreviation.Name) < 1 {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	if len(abbreviation.Description) < 1 {
		http.Error(w, "Description is required", http.StatusBadRequest)
		return
	}

	abbreviation.Name = strings.ToLower(abbreviation.Name)
	recordNr, err := persistence.AddAbbreviation(abbreviation)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var response = model.ResponseMessage{Message: "added: " + strconv.Itoa(recordNr)}
	responseJson, err := json.Marshal(response)

	helpers.WriteResponse(w, string(responseJson))
}
