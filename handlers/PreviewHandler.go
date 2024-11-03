package handlers

import (
	"fmt"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"log"
	"net/http"
	"swiki/helpers"
	"swiki/model"
	"text/template"
)

func PreviewHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	log.Println("PreviewHandler")
	content := r.PathValue("mdcontent")

	fmt.Println(content)

	tmpl, err := template.ParseFiles("templates/pagepreview.html")
	if tmpl == nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//md := []byte(page.Content)
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)
	result := string(markdown.ToHTML([]byte(content), nil, renderer))
	page := model.Page{Title: "Preview", Content: result}
	err = tmpl.Execute(w, page)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
