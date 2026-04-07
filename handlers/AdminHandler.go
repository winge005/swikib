package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"swiki/helpers"
	"swiki/model"
	"swiki/persistencelocal"
)

func AdminHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)

	position := strings.LastIndex(strings.ToLower(r.Pattern), "/")
	chosenAction := r.Pattern[position+1:]

	if chosenAction == "admin-updatefulldb" {
		var response = model.ResponseMessage{Message: "gonna update database"}
		responseJson, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		helpers.WriteResponse(w, string(responseJson))
		go persistencelocal.UpdateSync()
		return
	} else if chosenAction == "admin-syncfulldb" {
		var response = model.ResponseMessage{Message: "gonna sync full"}
		responseJson, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		helpers.WriteResponse(w, string(responseJson))
		go persistencelocal.Sync()
		return
	} else {
		http.Error(w, errors.New("Not supported").Error(), http.StatusInternalServerError)
	}

}
