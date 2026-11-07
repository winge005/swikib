package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"swiki/helpers"
	"swiki/model"
	"swiki/persistencelocal"
)

func AdminHandler(w http.ResponseWriter, r *http.Request) {

	position := strings.LastIndex(strings.ToLower(r.Pattern), "/")
	chosenAction := r.Pattern[position+1:]

	var run func()
	var message string

	switch chosenAction {
	case "admin-updatefulldb":
		run = persistencelocal.UpdateSync
		message = "gonna update database"
	case "admin-syncfulldb":
		run = persistencelocal.Sync
		message = "gonna sync full"
	default:
		http.Error(w, errors.New("Not supported").Error(), http.StatusInternalServerError)
		return
	}

	if !persistencelocal.TryStartProgress(chosenAction) {
		http.Error(w, "Er loopt al een proces", http.StatusConflict)
		return
	}

	responseJson, err := json.Marshal(model.ResponseMessage{Message: message})
	if err != nil {
		persistencelocal.FinishProgress(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	helpers.WriteResponse(w, string(responseJson))

	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				persistencelocal.FinishProgress(fmt.Errorf("%v", rec))
			}
		}()

		run()
		persistencelocal.FinishProgress(nil)
	}()
}
