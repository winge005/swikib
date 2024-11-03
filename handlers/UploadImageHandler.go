package handlers

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"swiki/helpers"
	"swiki/persistence"
)

func UploadImageHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}

	fileName := fileHeader.Filename
	if !checkValidExtension(fileName) {
		http.Error(w, "{errorText: 'File must be an image'}", http.StatusBadRequest)
		return
	}

	defer file.Close()

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println(fileName)

	result := persistence.AddImage(fileName, buf.Bytes())

	if !result {
		http.Error(w, "{errorText: 'saving image has failed'}", http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Upload succeeded")

}

func checkValidExtension(name string) bool {
	allowedExtensions := [9]string{"png", "apng", "jpg", "jpeg", "jfif", "jpeg", "gif", "webp", "svg"}
	allowed := false

	posPoint := strings.LastIndex(name, ".")
	name = name[posPoint+1:]
	name = strings.ToLower(name)

exit:
	for _, item := range allowedExtensions {
		if item == name {
			allowed = true
			break exit
		}
	}

	return allowed
}
