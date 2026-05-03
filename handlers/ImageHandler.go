package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"swiki/helpers"
	"swiki/persistence"
)

func ImageHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)

	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	imageKeys, ok := r.URL.Query()["image"]
	if !ok || len(imageKeys) == 0 {
		http.Error(w, "image parameter missing", http.StatusBadRequest)
		return
	}

	imageData := persistence.GetImageFrom(imageKeys[0])
	if imageData == nil {
		http.Error(w, "image not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Cache-Control", "public, max-age=604800") // 7 days
	_, err := w.Write(imageData)
	if err != nil {
		log.Println("writing file to browser failed: " + imageKeys[0])
		return
	}
}

func ImageSizeHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)

	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	err, picturesInfo := persistence.Get25biggestImages()
	if err != nil {
		log.Println("getting biggest images failed")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responseJson, err := json.Marshal(picturesInfo)
	helpers.WriteResponse(w, string(responseJson))
}
