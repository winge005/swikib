package handlers

import (
	"net/http"
	"swiki/helpers"
)

func PagePreflightHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
	return
}
