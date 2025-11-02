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

	imageKeys, _ := r.URL.Query()["image"]
	imageData := persistence.GetImageFrom(imageKeys[0])
	_, err := w.Write(imageData)
	if err != nil {
		log.Println("writing file to browser failed: " + imageKeys[0])
		log.Fatal(err)
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
