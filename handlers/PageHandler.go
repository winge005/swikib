package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"swiki/helpers"
	"swiki/model"
	"swiki/persistence"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
)

func PageAddHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	bd, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("%s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var page model.Page
	err = json.Unmarshal(bd, &page)

	if len(page.Title) == 0 {
		http.Error(w, "title is not filled", http.StatusInternalServerError)
		return
	}

	if len(page.Category) == 0 {
		http.Error(w, "category is not filled", http.StatusInternalServerError)
		return
	}

	if len(page.Content) == 0 {
		http.Error(w, "content is not filled", http.StatusInternalServerError)
		return
	}

	page.Category = strings.ToLower(page.Category)

	page.Title = strings.TrimSpace(page.Title)
	page.Content = strings.TrimSpace(page.Content)

	if persistence.IsPageTitleUsed(page.Title) {
		http.Error(w, "Title already used", http.StatusInternalServerError)
		return
	}
	id, err := persistence.AddPage(page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response = model.ResponseMessage{Message: "added: " + strconv.Itoa(id)}
	responseJson, err := json.Marshal(response)

	helpers.WriteResponse(w, string(responseJson))
}

func PageUpdateHandler(w http.ResponseWriter, r *http.Request) {
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

	bd, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("%s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var page model.Page
	err = json.Unmarshal(bd, &page)

	page.Id = idInt
	page.Category = strings.ToLower(page.Category)

	err = persistence.UpdatePage(page)
	if err != nil {
		log.Printf("%s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	helpers.WriteResponse(w, "")
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

	var response = model.ResponseMessage{Message: "Deleted: " + strconv.Itoa(idInt)}
	responseJson, err := json.Marshal(response)

	helpers.WriteResponse(w, string(responseJson))
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

	response, err := json.Marshal(page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	helpers.WriteResponse(w, string(response))
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

	response, err := json.Marshal(pages)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	helpers.WriteResponse(w, string(response))
}

func PageHandlerGetCategories(w http.ResponseWriter, r *http.Request) {
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

	var response []string

	for _, category := range categories {
		response = append(response, category)
	}

	responseJson, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	helpers.WriteResponse(w, string(responseJson))
}

func PrePageAddHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	bd, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("%s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var prePage model.PrePage
	err = json.Unmarshal(bd, &prePage)

	if len(prePage.Url) == 0 {
		http.Error(w, "title is not filled", http.StatusInternalServerError)
		return
	}

	err = persistence.AddPrePage(prePage.Url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response = model.ResponseMessage{Message: "added"}
	responseJson, err := json.Marshal(response)

	helpers.WriteResponse(w, string(responseJson))
}

func PrePageGetAllHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	prePages, err := persistence.GetAllPrePages()

	responseJson, err := json.Marshal(prePages)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(responseJson) == 0 {
		helpers.WriteResponse(w, "")
		return
	}

	helpers.WriteResponse(w, string(responseJson))

}

func PrePageDeleteHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	id := r.PathValue("id")
	idNumber, _ := strconv.Atoi(id)

	err := persistence.DeletePrePage(idNumber)
	if err != nil {
		log.Printf("%s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response = model.ResponseMessage{Message: "Deleted: " + id}
	responseJson, err := json.Marshal(response)

	helpers.WriteResponse(w, string(responseJson))
}

func PageAlreadyUsedHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	title := r.PathValue("title")

	used := false
	title = strings.TrimSpace(title)

	if persistence.IsPageTitleUsed(title) {
		used = true
	}

	var response = model.ResponseMessage{Message: strconv.FormatBool(used)}
	responseJson, err := json.Marshal(response)
	if err != nil {
		log.Printf("%s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	helpers.WriteResponse(w, string(responseJson))
}
