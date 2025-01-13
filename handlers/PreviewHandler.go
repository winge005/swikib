package handlers

import (
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"io"
	"log"
	"net/http"
	"strings"
	"swiki/helpers"
	"text/template"
)

func PreviewHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	log.Println("PreviewHandler")

	//var page model.Page
	//r.ParseForm()

	//content := r.PathValue("mdcontent")

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	bodyString = bodyString[111:]
	posTrailer := strings.Index(bodyString, "-----------------------------")
	bodyString = bodyString[:posTrailer]
	bodyString = strings.TrimSpace(bodyString)

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)
	result := string(markdown.ToHTML([]byte(bodyString), nil, renderer))

	tmpl, err := template.ParseFiles("templates/pagepreview.html")
	if tmpl == nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		PreviewContent string
	}{
		PreviewContent: result,
	}
	err = tmpl.Execute(w, data)
	return
}
