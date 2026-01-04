package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"swiki/handlers"
	"swiki/model"
	"swiki/persistence"
	"swiki/search"
	"syscall"

	"gopkg.in/yaml.v2"
)

var config = &model.Config{}

func main() {

	getConfig()
	persistence.SetConfig(config.Server.AccessToken)
	persistence.CreateTables()

	// pages
	http.HandleFunc("GET /swiki/pages/categories", handlers.PageHandlerGetCategories)
	http.HandleFunc("GET /swiki/pages/categories/{category}", handlers.PageCategoriesHandler)
	http.HandleFunc("GET /swiki/pages/{id}", handlers.PageViewHandler)
	http.HandleFunc("GET /swiki/pages/edit/{id}", handlers.PageEditHandler)
	http.HandleFunc("PUT /swiki/pages/{id}", handlers.PageUpdateHandler)
	http.HandleFunc("DELETE /swiki/pages/{id}", handlers.PageDeleteHandler)
	http.HandleFunc("POST /swiki/pages", handlers.PageAddHandler)
	http.HandleFunc("POST /swiki/page/preview", handlers.PreviewHandler)
	http.HandleFunc("POST /swiki/pages/uploadImage", handlers.UploadImageHandler)
	http.HandleFunc("GET /swiki/pages/created", handlers.SpecialsHandler)
	http.HandleFunc("GET /swiki/pages/exist/{title}", handlers.PageAlreadyUsedHandler)

	// pictures
	http.HandleFunc("GET /swiki/pages/imagessize", handlers.ImageSizeHandler)
	http.HandleFunc("GET /swiki/pages/image", handlers.ImageHandler)

	// page Statistics
	http.HandleFunc("GET /swiki/pages/statistics", handlers.PageGetStatistics)

	// prepages
	http.HandleFunc("GET /swiki/prepages", handlers.PrePageGetAllHandler)
	http.HandleFunc("DELETE /swiki/prepages/{id}", handlers.PrePageDeleteHandler)
	http.HandleFunc("POST /swiki/prepages", handlers.PrePageAddHandler)

	// links
	http.HandleFunc("GET /swiki/links/categories", handlers.LinksGetCategoriesHandlerr)
	http.HandleFunc("GET /swiki/links/categoriesoptions", handlers.LinksCategoryOptionsHandler)
	http.HandleFunc("GET /swiki/links/categorie/{category}", handlers.LinksFromCategoryHandler)
	http.HandleFunc("GET /swiki/links/update/{id}", handlers.LinkEditeHandler)
	http.HandleFunc("PUT /swiki/links/{id}", handlers.LinkUpdateHandler)
	http.HandleFunc("POST /swiki/links", handlers.LinkAddHandler)
	http.HandleFunc("GET /swiki/links/{id}", handlers.LinkViewHandler)
	http.HandleFunc("DELETE /swiki/links/{id}", handlers.LinkDeleteHandler)

	// abbreviations
	http.HandleFunc("POST /swiki/abbreviation", handlers.AbbreviationAddHandler)
	http.HandleFunc("GET /swiki/abbreviation/{fl}", handlers.AbbreviationHandler)
	http.HandleFunc("DELETE /swiki/abbreviation/{fl}", handlers.AbbreviationDelete)

	// search
	http.HandleFunc("POST /swiki/search", handlers.SearchHandler)

	// options
	http.HandleFunc("OPTIONS /swiki/pages/{id}", handlers.PagePreflightHandler)
	http.HandleFunc("OPTIONS /swiki/pages", handlers.PagePreflightHandler)
	http.HandleFunc("OPTIONS /swiki/links/{id}", handlers.PagePreflightHandler)
	http.HandleFunc("OPTIONS /swiki/links", handlers.PagePreflightHandler)
	http.HandleFunc("OPTIONS /swiki/prepages", handlers.PagePreflightHandler)
	http.HandleFunc("OPTIONS /swiki/prepages/{id}", handlers.PagePreflightHandler)
	http.HandleFunc("OPTIONS /swiki/abbreviation", handlers.PagePreflightHandler)
	http.HandleFunc("OPTIONS /swiki/abbreviation/{fl}", handlers.PagePreflightHandler)
	http.HandleFunc("OPTIONS /swiki/pages/exist/{title}", handlers.PagePreflightHandler)
	http.HandleFunc("OPTIONS /swiki/pages/statistics", handlers.PagePreflightHandler)
	http.HandleFunc("OPTIONS /swiki/search", handlers.PagePreflightHandler)

	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/", fs)

	// go images.CheckUneededImages()

	// persistence.Play()

	go search.CreateIndex(false)

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
