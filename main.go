package main

import (
	"context"
	"gopkg.in/yaml.v2"
	"log"
	"net/http"
	"os"
	"os/signal"
	"swiki/handlers"
	"swiki/model"
	"swiki/persistence"
	"syscall"
)

var config = &model.Config{}

func main() {

	getConfig()
	persistence.SetConfig(config.Server.AccessToken)
	persistence.CreateTables()

	http.HandleFunc("GET /swiki/page/categories", handlers.PageHandler)
	http.HandleFunc("GET /swiki/page/categoriesasoptions", handlers.PageCategoriesAsOptionsHandler)
	http.HandleFunc("GET /swiki/page/categories/{category}", handlers.PageCategoriesHandler)
	http.HandleFunc("GET /swiki/page/{id}", handlers.PageViewHandler)
	http.HandleFunc("GET /swiki/page/update/{id}", handlers.PageEditHandler)
	http.HandleFunc("PUT /swiki/page/{id}", handlers.PageUpdateHandler)
	http.HandleFunc("DELETE /swiki/page/{id}", handlers.PageDeleteHandler)
	http.HandleFunc("POST /swiki/page", handlers.PageAddHandler)
	http.HandleFunc("POST /swiki/page/preview", handlers.PreviewHandler)
	http.HandleFunc("GET /swiki/pages/image/", handlers.ImageHandler)
	http.HandleFunc("POST /swiki/pages/uploadImage", handlers.UploadImageHandler)

	http.HandleFunc("GET /swiki/links/categories", handlers.LinksGetCategoriesHandlerr)
	http.HandleFunc("GET /swiki/links/categoriesoptions", handlers.LinksCategoryOptionsHandler)
	http.HandleFunc("GET /swiki/links/categorie/{category}", handlers.LinksFromCategoryHandler)
	http.HandleFunc("GET /swiki/links/update/{id}", handlers.LinkEditeHandler)
	http.HandleFunc("PUT /swiki/links/update/{id}", handlers.LinkUpdateHandler)
	http.HandleFunc("POST /swiki/links", handlers.LinkAddHandler)
	http.HandleFunc("DELETE /swiki/links/{id}", handlers.LinkDeleteHandler)

	http.HandleFunc("POST /swiki/abbreviation", handlers.AbbreviationAddHandler)
	http.HandleFunc("GET /swiki/abbreviation/{fl}", handlers.AbbreviationHandler)

	http.HandleFunc("POST /swiki/search", handlers.SearchHandler)

	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/", fs)

	// go images.CheckUneededImages()
	// persistence.Play()

	srv := http.Server{}
	sigs := make(chan os.Signal, 1)
	done := make(chan struct{})

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Println("got interruption signal")
		if err := srv.Shutdown(context.TODO()); err != nil {
			log.Printf("server shutdown returned an err: %!v\n", err)
		}
		close(done)
		os.Exit(0)
	}()

	persistence.PerformCache()

	log.Printf("listening on port: %v\n", config.Server.Port)

	if err := http.ListenAndServe(":"+config.Server.Port, nil); err != nil {
		log.Fatal(err)
	}
}

func getConfig() {
	file, err := os.Open("config.yml")
	if err != nil {
		panic(err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)
	d := yaml.NewDecoder(file)
	if err := d.Decode(&config); err != nil {
		panic(err)
	}

	if len(config.Server.AccessToken) <= 0 {
		log.Println("Accesstoken: not available")
	}
}
