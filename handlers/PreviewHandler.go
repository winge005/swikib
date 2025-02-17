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

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	splittedString := strings.Split(bodyString, "\r\n")

	formId := strings.ReplaceAll(splittedString[0], "-", "")
	var content = ""
	for id, line := range splittedString {
		if id > 2 {
			if strings.Index(line, formId) != -1 {
				break
			}
			content += line + "\r\n"
		}
	}
	log.Println(content)

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)
	result := string(markdown.ToHTML([]byte(content), nil, renderer))

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
