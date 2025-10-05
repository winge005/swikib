package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"swiki/helpers"
	"swiki/model"
	"swiki/persistence"
	"text/template"
)

func LinksGetCategoriesHandlerr(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	categories, err := persistence.GetLinkCategories()
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

func LinksFromCategoryHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	cat := r.PathValue("category")

	links, err := persistence.GetLinksFromCategory(cat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responseJson, err := json.Marshal(links)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	helpers.WriteResponse(w, string(responseJson))

}

func LinkViewHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	id := r.PathValue("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	link, err := persistence.GetLink(idInt)
	responseJson, err := json.Marshal(link)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	helpers.WriteResponse(w, string(responseJson))
}

func LinkEditeHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	id := r.PathValue("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	link, err := persistence.GetLink(idInt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNoContent)
		return
	}

	categorien, err := persistence.GetLinkCategories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	tmpl, err := template.ParseFiles("templates/linkedit.html")
	if tmpl == nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Link       model.Link
		Categorien []string
	}{
		Link:       link,
		Categorien: categorien,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func LinkDeleteHandler(w http.ResponseWriter, r *http.Request) {
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

	err = persistence.DeleteLink(idInt)
	if err != nil {
		log.Printf("%s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response = model.ResponseMessage{Message: "Deleted: " + strconv.Itoa(idInt)}
	responseJson, err := json.Marshal(response)

	helpers.WriteResponse(w, string(responseJson))

}

func LinksCategoryOptionsHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	categories, err := persistence.GetLinkCategories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/linkscategoriesoptions.html")
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

func LinkAddHandler(w http.ResponseWriter, r *http.Request) {
	helpers.EnableCors(&w)
	if (*r).Method == http.MethodOptions {
		_, _ = w.Write([]byte("allowed"))
		return
	}

	bd, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("%s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var link model.Link
	err = json.Unmarshal(bd, &link)

	if len(link.Category) < 1 {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(link.Description) < 1 {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(link.Url) < 1 {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	link.Category = strings.ToLower(link.Category)

	id, err := persistence.AddLink(link)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response = model.ResponseMessage{Message: "added: " + strconv.Itoa(id)}
	responseJson, err := json.Marshal(response)

	helpers.WriteResponse(w, string(responseJson))
}

func LinkUpdateHandler(w http.ResponseWriter, r *http.Request) {
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var link model.Link
	err = json.Unmarshal(bd, &link)

	if len(link.Category) < 1 {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(link.Description) < 1 {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(link.Url) < 1 {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if link.Id < 1 || link.Id != idInt {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	link.Category = strings.ToLower(link.Category)

	err = persistence.UpdateLink(link)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	helpers.WriteResponse(w, "")
}
