package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
	"log"
	"os"
	"swiki/helpers"
	"swiki/model"
)

var accesskeyTurso string
var urlTurso string

var abbreviations = "CREATE TABLE IF NOT EXISTS abbreviations(id INTEGER PRIMARY KEY, name varchar(100) UNIQUE, description varchar(300))"
var links = "CREATE TABLE IF NOT EXISTS links(id INTEGER PRIMARY KEY, category varchar(100), url varchar(300) UNIQUE, description varchar(250), created varchar(19), updated varchar(19))"
var pages = "CREATE TABLE IF NOT EXISTS pages(id INTEGER, category varchar(200), title varchar(255), content TEXT, created varchar(19), updated varchar(19))"
var pictures = "CREATE TABLE IF NOT EXISTS pictures(id varchar(200), image BLOB, created varchar(19), updated varchar(19))"

var abbreviationsInCache = make(map[string][]model.Abbreviation)

func SetConfig(token string) {
	accesskeyTurso = token
	urlTurso = "libsql://swiki-winge005.turso.io?authToken=" + accesskeyTurso
}

func CreateTables() {
	database, err := sql.Open("libsql", urlTurso)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db %s: %s", urlTurso, err)
		os.Exit(1)
	}

	_, err = database.Exec(abbreviations)
	if err != nil {
		return
	}

	_, err = database.Exec(links)
	if err != nil {
		return
	}

	_, err = database.Exec(pages)
	if err != nil {
		return
	}

	_, err = database.Exec(pictures)
	if err != nil {
		return
	}
}

func GetCategories() ([]string, error) {
	var categories []string

	database, err := sql.Open("libsql", urlTurso)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db %s: %s", urlTurso, err)
		return categories, err
	}
	rows, err := database.Query("select DISTINCT category from pages order by category")
	if err != nil {
		log.Println(err.Error())
		return categories, err
	}

	defer rows.Close()
	defer database.Close()

	var category string

	for rows.Next() {
		err := rows.Scan(&category)
		if err != nil {
			log.Println(err.Error())
			return categories, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func GetPagesFromCategoryWithoutContent(whereCategory string) ([]model.Page, error) {
	var pages []model.Page

	database, err := sql.Open("libsql", urlTurso)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db %s: %s", urlTurso, err)
		return pages, err
	}

	rows, _ := database.Query("select id, title, category, created, updated from pages where category=?  order by title COLLATE NOCASE ASC", whereCategory)

	defer rows.Close()
	defer database.Close()

	var id int
	var title string
	var category string
	var created string
	var updated string

	for rows.Next() {
		err := rows.Scan(&id, &title, &category, &created, &updated)
		if err != nil {
			return nil, err
		}
		var page model.Page
		page.Id = id
		page.Title = title
		page.Category = category
		page.Content = ""
		page.Created = created
		page.Updated = updated
		pages = append(pages, page)
	}
	return pages, nil
}

func GetPage(idGiven int) (model.Page, error) {
	var page model.Page

	database, _ := sql.Open("libsql", urlTurso)
	rows, _ := database.Query("select id, category, title, content, created, updated from pages where id=?", idGiven)

	defer rows.Close()
	defer database.Close()

	var id int
	var category string
	var title string
	var content string
	var created string
	var updated string

	if rows == nil {
		return page, errors.New("no result")
	}

	for rows.Next() {
		err := rows.Scan(&id, &category, &title, &content, &created, &updated)
		if err != nil {
			return page, err
		}
		page.Id = id
		page.Category = category
		page.Title = title
		page.Content = content
		page.Created = created
		page.Updated = updated
		break
	}

	return page, nil
}

func GetPageFromOffset(recordTh int) (model.Page, error) {
	var page model.Page

	database, _ := sql.Open("libsql", urlTurso)
	rows, _ := database.Query("SELECT id, content FROM pages LIMIT 1 OFFSET ?;", recordTh)

	defer rows.Close()
	defer database.Close()

	var id int
	var content string

	if rows == nil {
		return page, errors.New("no result")
	}

	for rows.Next() {
		err := rows.Scan(&id, &content)
		if err != nil {
			return page, err
		}
		page.Id = id
		page.Content = content
		break
	}

	return page, nil
}

func IsPageTitleUsed(titleGiven string) bool {

	database, _ := sql.Open("libsql", urlTurso)
	rows, _ := database.Query("SELECT id, category FROM pages where title=?;", titleGiven)

	defer rows.Close()
	defer database.Close()

	var page model.Page
	var id int
	var category string

	if rows == nil {
		return false
	}

	for rows.Next() {
		err := rows.Scan(&id, &category)
		if err != nil {
			return false
		}
		page.Id = id
		page.Category = category
	}

	if page.Id > 0 {
		log.Printf("Title already exist in category %v id=%v", page.Category, page.Id)
		return true
	}

	return false

}

func GetImageFrom(id string) []byte {
	response := make([]byte, 0)
	database, err := sql.Open("libsql", urlTurso)

	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db %s: %s", urlTurso, err)
		os.Exit(1)
	}

	rows, _ := database.Query("select image from pictures where id=?", id)

	defer rows.Close()
	defer database.Close()

	var image []byte

	for rows.Next() {
		err := rows.Scan(&image)
		if err != nil {
			return nil
		}
		response = image
	}
	return response
}

func UpdatePage(newPage model.Page) error {
	database, err := sql.Open("libsql", urlTurso)
	if err != nil {
		return err
	}

	stmt, err := database.Prepare("UPDATE pages SET category=?, title=?, content=?, Updated=? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(newPage.Category, newPage.Title, newPage.Content, helpers.GetCurrentDateTime(), newPage.Id)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if affected == 1 {
		return nil
	}
	return errors.New("Nothing updated")
}

func AddPage(page model.Page) (int, error) {
	database, err := sql.Open("libsql", urlTurso)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db %s: %s", urlTurso, err)
		os.Exit(1)
	}

	var id int
	var rows *sql.Rows

	rows, _ = database.Query("select id from pages where title=?", page.Title)
	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			return 0, err
		}

		if id > 0 {
			return 0, errors.New("Already exist.")
		}
	}

	stmt, err := database.Prepare("INSERT INTO pages (title, category, content, created, updated) VALUES (?,?,?,?,?) RETURNING id")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	err = stmt.QueryRow(page.Title, page.Category, page.Content, helpers.GetCurrentDateTime(), "").Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func GetPagesCount() (int, error) {
	database, err := sql.Open("libsql", urlTurso)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db %s: %s", urlTurso, err)
		os.Exit(1)
	}
	defer database.Close()

	rows, err := database.Query("SELECT COUNT() as count FROM Pages")
	if err != nil {
		fmt.Printf("%s", err)
	}
	defer rows.Close()

	var count int

	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			return count, err
		}
	}

	return count, nil
}

func GetPageByQuery(queryPart string) ([]model.Page, error) {
	var page model.Page
	var pages []model.Page

	database, err := sql.Open("libsql", urlTurso)
	if err != nil {
		return pages, err
	}
	fmt.Println("select id, category, title, content, created, updated from pages where" + queryPart)
	rows, _ := database.Query("select id, category, title, content, created, updated from pages where" + queryPart)

	defer database.Close()

	var id int
	var category string
	var title string
	var content string
	var created string
	var updated string

	if rows == nil {
		return pages, nil
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&id, &category, &title, &content, &created, &updated)
		if err != nil {
			return pages, err
		}
		page.Id = id
		page.Category = category
		page.Title = title
		page.Content = content
		page.Created = created
		page.Updated = updated
		pages = append(pages, page)
	}

	return pages, nil
}

func AddImage(id string, data []byte) bool {
	database, err := sql.Open("libsql", urlTurso)

	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db %s: %s", urlTurso, err)
		os.Exit(1)
	}

	stmt, err := database.Prepare("INSERT INTO pictures (id, image, created, updated) VALUES (?,?,?,?)")
	defer database.Close()
	if err != nil {
		log.Println(err)
		return false
	}

	res, err := stmt.Exec(id, data, helpers.GetCurrentDateTime(), "")
	if err != nil {
		log.Println(err)
		return false
	}

	affected, err := res.RowsAffected()
	if affected != 1 {
		log.Println(err)
		return false
	}

	return true
}

func DeleteImage(id string) (int64, error) {
	database, err := sql.Open("libsql", urlTurso)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db %s: %s", urlTurso, err)
		os.Exit(1)
	}
	defer database.Close()

	stmt, err := database.Prepare("delete from pictures where id=?")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(id)
	if err != nil {
		return 0, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	return affected, err
}

func DeletePage(idGiven int) error {

	database, err := sql.Open("libsql", urlTurso)
	if err != nil {
		return err
	}

	stmt, err := database.Prepare("delete from pages where id=?")
	if err != nil {
		return err
	}
	res, err := stmt.Exec(idGiven)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 1 {
		return nil
	}

	return err
}

func GetLinkCategories() ([]string, error) {
	var categories []string
	database, err := sql.Open("libsql", urlTurso)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db %s: %s", urlTurso, err)
		return categories, err
	}
	rows, _ := database.Query("select DISTINCT category from links order by category")

	defer rows.Close()
	defer database.Close()

	var category string

	for rows.Next() {
		err := rows.Scan(&category)
		if err != nil {
			return categories, nil
		}
		categories = append(categories, category)
	}

	return categories, nil
}

func GetLinksFromCategory(whereCategory string) ([]model.Link, error) {
	var links []model.Link

	database, err := sql.Open("libsql", urlTurso)
	if err != nil {
		return links, err
	}
	rows, _ := database.Query("select id, url, description, category, created, IFNULL(updated, '') from links where category=?  order by description COLLATE NOCASE ASC", whereCategory)

	defer rows.Close()
	defer database.Close()

	var id int
	var url string
	var description string
	var category string
	var created string
	var updated string

	for rows.Next() {
		err := rows.Scan(&id, &url, &description, &category, &created, &updated)
		if err != nil {
			return links, err
		}
		var link model.Link
		link.Id = id
		link.Url = url
		link.Category = category
		link.Description = description
		link.Created = created
		link.Updated = updated
		links = append(links, link)
	}
	return links, nil
}

func GetLink(idGiven int) (model.Link, error) {
	var link model.Link
	database, err := sql.Open("libsql", urlTurso)
	if err != nil {
		return link, err
	}

	rows, err := database.Query("select category, url, description, created, IFNULL(updated, '') from links where id=?", idGiven)

	if err != nil {
		return link, err
	}

	defer rows.Close()
	defer database.Close()

	var category string
	var url string
	var description string
	var created string
	var updated string

	if rows == nil {
		return link, err
	}

	if rows.Next() {
		err := rows.Scan(&category, &url, &description, &created, &updated)
		if err != nil {
			return link, err
		}

		link.Id = idGiven
		link.Category = category
		link.Url = url
		link.Description = description
		link.Created = created
		link.Updated = updated
	}

	return link, nil
}

func LinkExist(givenUrl string) bool {

	database, err := sql.Open("libsql", urlTurso)
	if err != nil {
		return false
	}
	rows, _ := database.Query("select id from links where url = ?", givenUrl)

	defer rows.Close()
	defer database.Close()

	var id int

	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			return false
		}
		return true
	}
	return false
}

func UpdateLink(newLink model.Link) error {
	database, err := sql.Open("libsql", urlTurso)
	if err != nil {
		return err
	}

	stmt, err := database.Prepare("UPDATE links SET category=?, url=?, description=?, Updated=? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(newLink.Category, newLink.Url, newLink.Description, helpers.GetCurrentDateTime(), newLink.Id)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if affected == 1 {
		return nil
	}
	return errors.New("Nothing updated")
}

func AddLink(newLink model.Link) (int, error) {
	database, err := sql.Open("libsql", urlTurso)
	if err != nil {
		return 0, err
	}

	var id int
	var rows *sql.Rows

	rows, _ = database.Query("select id from links where url=?", newLink.Url)
	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			return 0, err
		}

		if id > 0 {
			return 0, errors.New("Already exist.")
		}
	}

	stmt, err := database.Prepare("INSERT INTO links(category, url, description, created) VALUES (?,?,?,?) RETURNING id")
	if err != nil {
		return 0, err
	}

	defer stmt.Close()

	err = stmt.QueryRow(newLink.Category, newLink.Url, newLink.Description, helpers.GetCurrentDateTime()).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func DeleteLink(idGiven int) error {
	database, err := sql.Open("libsql", urlTurso)

	if err != nil {
		return err
	}

	stmt, err := database.Prepare("delete from links where id=?")
	if err != nil {
		return err
	}
	_, err = stmt.Exec(idGiven)
	if err != nil {
		return err
	}

	return nil
}

func PerformCache() {
	setAbbreviationInCache()
}

func setAbbreviationInCache() {
	var abbreviation model.Abbreviation
	database, _ := sql.Open("libsql", urlTurso)
	rows, _ := database.Query("select id, name, description from abbreviations")
	defer rows.Close()
	defer database.Close()

	var id int
	var name string
	var description string

	for rows.Next() {
		err := rows.Scan(&id, &name, &description)
		if err != nil {
			log.Panic("No rows found for abbreviations")
		}
		abbreviation.Id = id
		abbreviation.Name = name
		abbreviation.Description = description
		firstLetter := string(abbreviation.Name[0])

		_, ok := abbreviationsInCache[firstLetter]
		if ok {
			abbreviationsInCache[firstLetter] = append(abbreviationsInCache[firstLetter], abbreviation)
		} else {
			var abbreviations []model.Abbreviation
			abbreviations = append(abbreviations, abbreviation)
			abbreviationsInCache[firstLetter] = abbreviations
		}
	}
}

func GetAbbreviationsForLetter(givenLetter string) ([]model.Abbreviation, error) {
	return abbreviationsInCache[givenLetter], nil
}

func GetImageCount() (int, error) {
	database, err := sql.Open("libsql", urlTurso)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db %s: %s", urlTurso, err)
		os.Exit(1)
	}
	defer database.Close()

	rows, err := database.Query("SELECT COUNT() as count FROM pictures")

	if err != nil {
		fmt.Printf("%s", err)
	}
	defer rows.Close()

	var count int

	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			return count, err
		}
	}

	return count, nil
}

func GetImageFromOffSet(recordTh int) (string, error) {
	var image string

	database, _ := sql.Open("libsql", urlTurso)
	rows, _ := database.Query("SELECT id FROM pictures LIMIT 1 OFFSET ?;", recordTh)

	defer rows.Close()
	defer database.Close()

	var id string

	if rows == nil {
		return image, errors.New("no result")
	}

	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			return image, err
		}
		image = id

		break
	}

	return image, nil
}

func Play() {

	//var pages []model.Page
	//var page model.Page

	database, err := sql.Open("libsql", urlTurso)
	if err != nil {
		log.Println("not connected to db")
		return
	}

	rows, err := database.Query("select id, category,description, url from links where id='245';")
	if err != nil {
		log.Println(err.Error())
		return
	}
	//stmt, err := database.Prepare("select id, category,title  from links where category='';")

	defer database.Close()
	//defer stmt.Close()
	defer rows.Close()

	//_, err = stmt.Exec()
	//if err != nil {
	//	log.Println("failed deleting")
	//	return
	//}

	var id int
	var category string
	var description string
	var url string
	//var content string
	//var created string
	//var updated string
	//
	for rows.Next() {
		err := rows.Scan(&id, &category, &description, &url)
		if err != nil {
			return
		}
		fmt.Println(id, category, description, url)

		//	page.Content = content
		//	page.Created = created
		//	page.Updated = updated
		//	pages = append(pages, page)
	}

	UpdateLink(model.Link{Id: id, Category: "go", Url: url, Description: description})

	//for _, page := range pages {
	//fmt.Println(page.Id)
	//fmt.Println(page.Title)
	//fmt.Println(page.Category)
	//fmt.Println(page.Content)
	//fmt.Println(page.Created)
	//fmt.Println(page.Updated)
	//DeletePage(1873)
	//}
	fmt.Println("Prima de luxe")
}
