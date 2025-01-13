package handlers

import (
	"log"
	"net/http"
	"strconv"
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

	tmpl, err := template.ParseFiles("templates/linkscategories.html")
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

	tmpl, err := template.ParseFiles("templates/linkslist.html")
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, links)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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

	http.Redirect(w, r, "http://localhost:5001/links.html", http.StatusSeeOther)

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

	log.Println("LinkAddHandler")

	var link model.Link

	r.ParseForm()

	description := r.FormValue("description")
	category := r.FormValue("category")
	newCategory := r.FormValue("newcategory")
	url := r.FormValue("url")

	link.Description = description
	link.Url = url
	if category == "Select category" {
		category = ""
	}

	var errorMessage = ""
	if len(category) > 0 && len(newCategory) > 0 {
		errorMessage = "category and newCategory may not be filled both\n"
	}

	if len(description) == 0 {
		errorMessage = "description is not filled\n"
	}

	if len(url) == 0 {
		errorMessage = "Url must be filled\n"
	}

	if len(category) > 0 {
		link.Category = category
	}
	if len(newCategory) > 0 {
		link.Category = newCategory
	}

	tmpl, err := template.ParseFiles("templates/linkaddresponse.html")
	if err != nil || len(errorMessage) > 0 {
		if err != nil {
			errorMessage = err.Error()
		}
		data := struct {
			Link         model.Link
			ErrorMessage string
			Message      string
		}{
			Link:         link,
			ErrorMessage: errorMessage,
			Message:      "Something wrong with the template",
		}
		err = tmpl.Execute(w, data)
		return
	}

	if persistence.LinkExist(link.Url) {
		errorMessage = "Link already exists"
		data := struct {
			Link         model.Link
			ErrorMessage string
			Message      string
		}{
			Link:         link,
			ErrorMessage: errorMessage,
			Message:      "",
		}
		err = tmpl.Execute(w, data)
		return
	}

	if len(errorMessage) == 0 {
		addedRecordId, err := persistence.AddLink(link)
		if err != nil {
			errorMessage = err.Error()
			data := struct {
				Link         model.Link
				ErrorMessage string
				Message      string
			}{
				Link:         link,
				ErrorMessage: err.Error(),
				Message:      "",
			}
			err = tmpl.Execute(w, data)
			return
		}
		data := struct {
			Link         model.Link
			ErrorMessage string
			Message      string
		}{
			Link:         link,
			ErrorMessage: "",
			Message:      "Added link Id: " + strconv.Itoa(addedRecordId),
		}
		err = tmpl.Execute(w, data)
	}
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

	oldLink, err := persistence.GetLink(idInt)
	if err != nil {
		log.Printf("%s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	r.ParseForm()

	if len(r.FormValue("category")) > 0 && len(r.FormValue("newcategory")) > 0 {
		log.Printf("%s", "category and newCategory can't be filled both")
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	var category string
	if len(r.FormValue("newcategory")) > 0 {
		category = r.FormValue("newcategory")
	} else {
		category = r.FormValue("category")
	}

	intId, err := strconv.Atoi(id)
	if err != nil {
		log.Printf("%s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newLink := model.Link{Id: intId, Description: r.FormValue("description"), Url: r.FormValue("url"), Created: oldLink.Created, Category: category}

	if newLink.Description == oldLink.Description && newLink.Url == oldLink.Url && newLink.Category == oldLink.Category {
		helpers.WriteResponse(w, "Nothing updated, all the same")
	}

	persistence.UpdateLink(newLink)

	helpers.WriteResponse(w, "<a hx-get=\"/swiki/links/categorie/"+category+"\" hx-swap=\"innerHTML\">Show list</a> ")
}
