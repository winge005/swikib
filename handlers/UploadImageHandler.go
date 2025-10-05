package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"path/filepath"
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

	uploaded := make(map[string]string)

	if r.Method != http.MethodPost {
		http.Error(w, "only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("Content-Type: %s", r.Header.Get("Content-Type"))

	const maxUpload = 64 << 20 // 64MB for testing
	r.Body = http.MaxBytesReader(w, r.Body, maxUpload)

	if err := r.ParseMultipartForm(maxUpload); err != nil {
		http.Error(w, "ParseMultipartForm error: "+err.Error(), http.StatusBadRequest)
		return
	}
	if r.MultipartForm == nil {
		http.Error(w, "no multipart form present", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, "no files uploaded under key 'files'", http.StatusBadRequest)
		return
	}

	for _, fh := range files {
		file, err := fh.Open()
		if err != nil {
			http.Error(w, "cannot open part: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		name := filepath.Base(fh.Filename)
		if !checkValidExtension(name) {
			uploaded[name] = "extension not allowed"
			continue
		}

		buf := bytes.NewBuffer(nil)
		if _, err := io.Copy(buf, file); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		result := persistence.AddImage(name, buf.Bytes())
		if !result {
			http.Error(w, "{errorText: 'saving image has failed'}", http.StatusBadRequest)
			return
		}

		uploaded[name] = "ok"
	}

	w.Header().Set("Content-Type", "application/json")
	response, err := json.Marshal(uploaded)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	helpers.WriteResponse(w, string(response))

	//file, fileHeader, err := r.FormFile("files")
	//if err != nil {
	//	http.Error(w, err.Error(), http.StatusBadRequest)
	//	return
	//}
	//
	//if err := r.ParseForm(); err != nil {
	//	fmt.Fprintf(w, "ParseForm() err: %v", err)
	//	return
	//}
	//
	//fileName := fileHeader.Filename
	//if !checkValidExtension(fileName) {
	//	http.Error(w, "{errorText: 'File must be an image'}", http.StatusBadRequest)
	//	return
	//}
	//
	//defer file.Close()
	//
	//buf := bytes.NewBuffer(nil)
	//if _, err := io.Copy(buf, file); err != nil {
	//	http.Error(w, err.Error(), http.StatusInternalServerError)
	//	return
	//}
	//
	//log.Println(fileName)
	//
	//result := persistence.AddImage(fileName, buf.Bytes())
	//
	//if !result {
	//	http.Error(w, "{errorText: 'saving image has failed'}", http.StatusBadRequest)
	//	return
	//}
	//
	//fmt.Fprintf(w, "Upload succeeded")
	//var response = model.ResponseMessage{Message: "Uploaded: " + fileName}
	//responseJson, err := json.Marshal(response)
	//
	//

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
