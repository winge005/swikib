package helpers

import (
	"fmt"
	"net/http"
)

func WriteErrorResponse(w http.ResponseWriter, response string, errorNummer int) {
	w.WriteHeader(errorNummer)
	_, _ = fmt.Fprint(w, response)
}
