package handlers

import (
	"fmt"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"swiki/helpers"
	"swiki/model"
	"swiki/persistence"
	"text/template"
	"time"
)

func PageHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	fmt.Printf("Request at %v\n", time.Now())
	for k, v := range r.Header {
		fmt.Printf("%v: %v\n", k, v)
	}

	categories, err := persistence.GetCategories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/categories.html")
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, categories)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func PageCategoriesAsOptionsHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	categories, err := persistence.GetCategories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/categoriesasoptions.html")
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, categories)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func PageAddHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	log.Println("PageAddHandler")

	var page model.Page
	r.ParseForm()

	category := r.FormValue("category")
	newCategory := r.FormValue("newcategory")
	title := r.FormValue("title")
	content := r.FormValue("mdcontent")
	tmpl, err := template.ParseFiles("templates/pageaddonsuccess.html")

	var errorMessage = ""
	if len(category) > 0 && len(newCategory) > 0 {
		errorMessage = "category and newCategory may not be filled both\n"
	}

	if len(category) == 0 && len(newCategory) == 0 {
		errorMessage = "category not be filled\n"
	}

	if len(title) == 0 {
		errorMessage = "title is not filled\n"
	}

	if len(content) == 0 {
		errorMessage = "Content must be filled\n"
	}

	page.Category = category
	if len(newCategory) > 0 {
		page.Category = newCategory
	}

	if len(errorMessage) == 0 {
		page.Title = title
		page.Content = content

		if persistence.IsPageTitleUsed(page.Title) {
			errorMessage = "Title already used"
			tmpl, err = template.ParseFiles("templates/pageaddonerror.html")
			data := struct {
				Page         model.Page
				ErrorMessage string
			}{
				Page:         page,
				ErrorMessage: errorMessage,
			}
			err = tmpl.Execute(w, data)
			return
		}

		id, err := persistence.AddPage(page)
		if err != nil {
			errorMessage = err.Error()
			tmpl, err = template.ParseFiles("templates/pageaddonerror.html")
			data := struct {
				Page         model.Page
				ErrorMessage string
			}{
				Page:         page,
				ErrorMessage: errorMessage,
			}
			err = tmpl.Execute(w, data)
			return
		}
		page.Id = 0
		page.Title = ""
		page.Category = ""
		page.Content = ""
		page.Created = ""
		page.Updated = ""
		data := struct {
			Page model.Page
			Id   int
		}{
			Page: page,
			Id:   id,
		}
		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl, err = template.ParseFiles("templates/pageaddonerror.html")
	page.Title = title
	page.Content = content

	data := struct {
		Page         model.Page
		ErrorMessage string
	}{
		Page:         page,
		ErrorMessage: errorMessage,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func PageUpdateHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	id := r.PathValue("id")
	bd, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("%s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	splittedData := strings.Split(string(bd), "&")
	var page model.Page
	idInt, err := strconv.Atoi(id)

	page.Id = idInt
	var oldCategory string

	for _, kv := range splittedData {
		kvs := strings.Split(kv, "=")
		if kvs[0] == "category" {
			if len(kvs[1]) > 0 {
				page.Category = helpers.RemoveUrlEncoding(kvs[1])
			}
		} else if kvs[0] == "newcategory" {
			if len(kvs[1]) > 0 {
				page.Category = helpers.RemoveUrlEncoding(kvs[1])
			}
		} else if kvs[0] == "title" {
			page.Title = helpers.RemoveUrlEncoding(kvs[1])
		} else if kvs[0] == "mdcontent" {
			page.Content = helpers.RemoveUrlEncoding(kvs[1])
		} else if kvs[0] == "oldcategory" {
			oldCategory = helpers.RemoveUrlEncoding(kvs[1])
		}
	}

	err = persistence.UpdatePage(page)
	if err != nil {
		log.Printf("%s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	returnVal := "<div class=\"alert alert-primary\" role=\"alert\">\n<button class=\"btn btn-outline-success\" type=\"button\" hx-target=\"#content\" hx-get=\"/swiki/page/categories/$oldcategory$\">\n              <svg xmlns=\"http://www.w3.org/2000/svg\" width=\"16\" height=\"16\" fill=\"currentColor\" class=\"bi bi-arrow-left-square\" viewBox=\"0 0 16 16\">\n                <path fill-rule=\"evenodd\" d=\"M15 2a1 1 0 0 0-1-1H2a1 1 0 0 0-1 1v12a1 1 0 0 0 1 1h12a1 1 0 0 0 1-1zM0 2a2 2 0 0 1 2-2h12a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2H2a2 2 0 0 1-2-2zm11.5 5.5a.5.5 0 0 1 0 1H5.707l2.147 2.146a.5.5 0 0 1-.708.708l-3-3a.5.5 0 0 1 0-.708l3-3a.5.5 0 1 1 .708.708L5.707 7.5z\"/>\n              </svg>\n            </button> Updated</div>\n"
	returnVal = strings.Replace(returnVal, "$oldcategory$", oldCategory, -1)
	helpers.WriteResponse(w, returnVal)
}

func PageEditHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	log.Println("PageEditHandler")
	id := r.PathValue("id")
	idGiven, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	page, err := persistence.GetPage(idGiven)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	categorien, err := persistence.GetCategories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	tmpl, err := template.ParseFiles("templates/pageedit.html")
	if tmpl == nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Page       model.Page
		Categorien []string
	}{
		Page:       page,
		Categorien: categorien,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func PageDeleteHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	id := r.PathValue("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		log.Printf("%s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = persistence.DeletePage(idInt)
	if err != nil {
		log.Printf("%s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "http://localhost:5001/pages.html", http.StatusSeeOther)
}

func PageViewHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	log.Println("PageViewHandler")
	id := r.PathValue("id")
	idGiven, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	page, err := persistence.GetPage(idGiven)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	md := []byte(page.Content)
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)
	page.Content = string(markdown.ToHTML(md, nil, renderer))
	imagesFound := helpers.ProcessImagesFromHtml(page.Content)
	if len(imagesFound) > 0 {
		replaceImageTags(imagesFound, &page.Content)
	}

	tmpl, err := template.ParseFiles("templates/pageview.html")
	if tmpl == nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Page model.Page
	}{
		Page: page,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// TODO: make this react on config
func replaceImageTags(imagesFound []string, content *string) {
	for _, v := range imagesFound {
		*content = strings.Replace(*content, v, "http://localhost:5001/swiki/pages/image?image="+v, -1)
		fmt.Println(v)
	}
}

func PageCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	category := r.PathValue("category")
	pages, err := persistence.GetPagesFromCategoryWithoutContent(category)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/pageslist.html")
	if tmpl == nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, pages)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("categories")
}

//412
