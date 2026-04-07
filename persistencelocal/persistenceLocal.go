package persistencelocal

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"swiki/helpers"
	"swiki/model"

	_ "modernc.org/sqlite"
)

var dbName = "wiki.db"
var driverName = "sqlite"

var abbreviations = "CREATE TABLE IF NOT EXISTS abbreviations(id INTEGER PRIMARY KEY, name varchar(100) UNIQUE, description varchar(300), tursoid INTEGER)"
var links = "CREATE TABLE IF NOT EXISTS links(id INTEGER PRIMARY KEY, category varchar(100), url varchar(300) UNIQUE, description varchar(250), created varchar(19), updated varchar(19), tursoid INTEGER)"
var pages = "CREATE TABLE IF NOT EXISTS pages(id INTEGER PRIMARY KEY, category varchar(200), title varchar(255), content TEXT, created varchar(19), updated varchar(19), tursoid INTEGER)"
var pictures = "CREATE TABLE IF NOT EXISTS pictures(id varchar(200), image BLOB, created varchar(19), updated varchar(19), tursoid varchar(200))"

func CreateTables() {
	database, err := sql.Open(driverName, dbName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open localdb %s: %s", dbName, err)
		os.Exit(1)
	}

	defer database.Close()

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

	database, err := sql.Open(driverName, dbName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db %s: %s", dbName, err)
		return categories, err
	}
	defer database.Close()

	rows, err := database.Query("select DISTINCT category from pages order by category")
	if err != nil {
		log.Println(err.Error())
		return categories, err
	}
	defer rows.Close()

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

func GetPagesFromCategoryWithContent(whereCategory string) ([]model.PageLocal, error) {
	var pages []model.PageLocal

	database, err := sql.Open(driverName, dbName)
	if err != nil {
		log.Println(err.Error())
		return pages, err
	}

	rows, _ := database.Query("select id, title, category, content, created, updated, tursoid from pages where category=?  order by title COLLATE NOCASE ASC", whereCategory)

	defer rows.Close()
	defer database.Close()

	var id int
	var title string
	var category string
	var content string
	var created string
	var updated string
	var tursoId int

	for rows.Next() {
		err := rows.Scan(&id, &title, &category, &content, &created, &updated, &tursoId)
		if err != nil {
			return nil, err
		}
		var page model.PageLocal
		page.Id = id
		page.Title = title
		page.Category = category
		page.Content = content
		page.Created = created
		page.Updated = updated
		page.TursoId = tursoId
		pages = append(pages, page)
	}
	return pages, nil
}

func GetPagesFromCategoryWithoutContent(whereCategory string) ([]model.PageLocal, error) {
	var pages []model.PageLocal

	database, err := sql.Open(driverName, dbName)
	if err != nil {
		log.Println(err.Error())
		return pages, err
	}

	rows, _ := database.Query("select id, title, created, updated, tursoid from pages where category=?  order by title COLLATE NOCASE ASC", whereCategory)

	defer rows.Close()
	defer database.Close()

	var id int
	var title string
	var created string
	var updated string
	var tursoid int

	for rows.Next() {
		err := rows.Scan(&id, &title, &created, &updated, &tursoid)
		if err != nil {
			return nil, err
		}
		var page model.PageLocal
		page.Id = id
		page.Title = title
		page.Content = ""
		page.Created = created
		page.Updated = updated
		page.TursoId = tursoid
		pages = append(pages, page)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return pages, nil
}

func GetPage(idGiven int) (model.PageLocal, error) {
	var page model.PageLocal

	database, err := sql.Open(driverName, dbName)
	if err != nil {
		log.Println(err.Error())
		return page, err
	}
	defer database.Close()
	rows, _ := database.Query("select id, category, title, content, created, updated, tursoid from pages where id=?", idGiven)

	defer rows.Close()

	var id int
	var category string
	var title string
	var content string
	var created string
	var updated string
	var tursoid int

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
		page.TursoId = tursoid
		break
	}

	return page, nil
}

func GetImageFrom(id string) []byte {
	response := make([]byte, 0)

	database, err := sql.Open(driverName, dbName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open localdb %s: %s", dbName, err)
		os.Exit(1)
	}

	rows, _ := database.Query("select image from pictures where id=?", id)

	defer database.Close()
	defer rows.Close()

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

func UpdatePage(updatedPage model.PageLocal) error {
	database, err := sql.Open(driverName, dbName)
	if err != nil {
		return err
	}
	defer database.Close()

	stmt, err := database.Prepare("UPDATE pages SET category=?, title=?, content=?, Updated=? WHERE tursoid = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	updatedPage.Category = strings.ToLower(updatedPage.Category)

	res, err := stmt.Exec(updatedPage.Category, updatedPage.Title, updatedPage.Content, updatedPage.Updated)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if affected == 1 {
		return nil
	}
	return errors.New("nothing updated")
}

func AddPage(page model.PageLocal) (int, error) {
	database, err := sql.Open(driverName, dbName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open localdb %s: %s", dbName, err)
		os.Exit(1)
	}
	defer database.Close()
	var id int

	stmt, err := database.Prepare("INSERT INTO pages (title, category, content, created, updated, tursoid) VALUES (?,?,?,?,?,?) RETURNING id")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(page.Title, page.Category, page.Content, helpers.GetCurrentDateTime(), "", page.TursoId).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func GetPagesCount() (int, error) {
	database, err := sql.Open(driverName, dbName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open localdb %s: %s", dbName, err)
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

func DeleteImage(id string) (int64, error) {
	database, err := sql.Open(driverName, dbName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db %s: %s", dbName, err)
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

func AddImage(picture model.Picture) bool {
	database, err := sql.Open(driverName, dbName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db %s: %s", dbName, err)
		os.Exit(1)
	}
	defer database.Close()

	stmt, err := database.Prepare("INSERT INTO pictures (id, image, created) VALUES (?,?,?)")
	defer stmt.Close()
	if err != nil {
		log.Println(err)
		return false
	}

	res, err := stmt.Exec(picture.Id, picture.ImageBytes, picture.Created)
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

func AddLink(newLink model.LinkLocal) (int, error) {
	database, err := sql.Open(driverName, dbName)
	if err != nil {
		return 0, err
	}
	defer database.Close()
	var id int

	stmt, err := database.Prepare("INSERT INTO links(category, url, description, created, tursoid) VALUES (?,?,?,?,?) RETURNING id")
	if err != nil {
		return 0, err
	}

	defer stmt.Close()

	err = stmt.QueryRow(newLink.Category, newLink.Url, newLink.Description, newLink.Created, newLink.TursoId).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func AddAbbreviation(abbreviation model.AbbreviationLocal) (int, error) {
	database, err := sql.Open(driverName, dbName)
	if err != nil {
		return 0, err
	}

	defer database.Close()

	var id int
	stmt, err := database.Prepare("INSERT INTO abbreviations (name, description, tursoid) VALUES (?,?,?) RETURNING id")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	err = stmt.QueryRow(strings.ToUpper(abbreviation.Name), abbreviation.Description, abbreviation.TursoId).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func GetAbbreviation(givenAbbreviation string) (model.AbbreviationLocal, error) {
	var abbreviation model.AbbreviationLocal
	database, err := sql.Open(driverName, dbName)
	if err != nil {
		return abbreviation, err
	}

	defer database.Close()
	rows, _ := database.Query("Select * from abbreviations where name=?", givenAbbreviation)
	defer rows.Close()

	var id int
	var name string
	var description string
	var tursoid int

	if rows == nil {
		return abbreviation, errors.New("no result")
	}

	for rows.Next() {
		err := rows.Scan(&id, &name, &description, &tursoid)
		if err != nil {
			return abbreviation, err
		}
		abbreviation.Id = id
		abbreviation.Name = name
		abbreviation.Description = description
		abbreviation.TursoId = tursoid
		break
	}

	return abbreviation, nil
}

func DropTable(tbl string) error {
	database, err := sql.Open(driverName, dbName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open localdb %s: %s", dbName, err)
		return err
	}
	defer database.Close()

	pages := fmt.Sprintf("DROP TABLE IF EXISTS %s;", tbl)

	_, err = database.Exec(pages)
	if err != nil {
		return err
	}

	return nil
}
