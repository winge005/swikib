package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"swiki/helpers"
	"swiki/model"
	"sync"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

var db *sql.DB
var accesskeyTurso string
var urlTurso string

var (
	abbreviationsInCache = make(map[string][]model.Abbreviation)
	abbrMu               sync.RWMutex
)

var abbreviations = "CREATE TABLE IF NOT EXISTS abbreviations(id INTEGER PRIMARY KEY, name varchar(100) UNIQUE, description varchar(300))"
var links = "CREATE TABLE IF NOT EXISTS links(id INTEGER PRIMARY KEY, category varchar(100), url varchar(300) UNIQUE, description varchar(250), created varchar(19), updated varchar(19))"
var pages = "CREATE TABLE IF NOT EXISTS pages(id INTEGER, category varchar(200), title varchar(255), content TEXT, created varchar(19), updated varchar(19))"
var pictures = "CREATE TABLE IF NOT EXISTS pictures(id varchar(200), image BLOB, created varchar(19), updated varchar(19))"
var prePages = "CREATE TABLE IF NOT EXISTS prepages(id INTEGER PRIMARY KEY, url varchar(300) UNIQUE, created varchar(19))"

func SetConfig(token string) {
	accesskeyTurso = token
	urlTurso = "libsql://swiki-winge005.turso.io?authToken=" + accesskeyTurso
}

func InitDB() error {
	var err error
	db, err = sql.Open("libsql", urlTurso)
	if err != nil {
		return fmt.Errorf("failed to open db %s: %w", urlTurso, err)
	}
	// Check connection
	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping db %s: %w", urlTurso, err)
	}
	return nil
}

func CreateTables() {
	if db == nil {
		if err := InitDB(); err != nil {
			log.Fatal(err)
		}
	}

	_, err := db.Exec(abbreviations)
	if err != nil {
		return
	}

	_, err = db.Exec(links)
	if err != nil {
		return
	}

	_, err = db.Exec(pages)
	if err != nil {
		return
	}

	_, err = db.Exec(pictures)
	if err != nil {
		return
	}

	_, err = db.Exec(prePages)
	if err != nil {
		return
	}
}

func GetCategories() ([]string, error) {
	var categories []string

	rows, err := db.Query("select DISTINCT category from pages order by category")
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

func GetCategoryCount(whereCategory string) (int, error) {

	rows, err := db.Query("SELECT COUNT(*) as count from pages WHERE category = ?", whereCategory)
	if err != nil {
		log.Println(err.Error())
		return 0, err
	}

	defer rows.Close()

	var count int

	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			log.Println(err.Error())
			return 0, err
		}
		return count, nil
	}

	return 0, nil
}

func GetPagesFromCategoryWithContent(whereCategory string) ([]model.Page, error) {
	var pages []model.Page

	rows, err := db.Query("select id, title, category, content, created, updated from pages where category=?  order by title COLLATE NOCASE ASC", whereCategory)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var id int
	var title string
	var category string
	var content string
	var created string
	var updated string

	for rows.Next() {
		err := rows.Scan(&id, &title, &category, &content, &created, &updated)
		if err != nil {
			return nil, err
		}
		var page model.Page
		page.Id = id
		page.Title = title
		page.Category = category
		page.Content = content
		page.Created = created
		page.Updated = updated
		pages = append(pages, page)
	}
	return pages, nil
}

func GetPagesFromDateAndAfterWithContent(afterDate string) ([]model.Page, error) {
	var pages []model.Page

	year := afterDate[6:10]
	month := afterDate[3:5]
	day := afterDate[0:2]
	hour := afterDate[11:13]
	minute := afterDate[14:16]
	second := afterDate[17:19]

	compareDate := year + "-" + month + "-" + day + " " + hour + ":" + minute + ":" + second

	rows, err := db.Query("SELECT id, title, category, content, Created, updated FROM pages WHERE substr(created, 7, 4) || '-' || substr(created, 4, 2) || '-' || substr(created, 1, 2) || ' ' || substr(created, 12) > ?;", compareDate)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var id int
	var title string
	var category string
	var content string
	var created string
	var updated string

	for rows.Next() {
		err := rows.Scan(&id, &title, &category, &content, &created, &updated)
		if err != nil {
			return nil, err
		}
		var page model.Page
		page.Id = id
		page.Title = title
		page.Category = category
		page.Content = content
		page.Created = created
		page.Updated = updated
		pages = append(pages, page)
	}
	return pages, nil
}

func GetUpdatedPagesFromDateAndAfterWithContent(afterDate string) ([]model.Page, error) {
	var pages []model.Page

	year := afterDate[6:10]
	month := afterDate[3:5]
	day := afterDate[0:2]
	hour := afterDate[11:13]
	minute := afterDate[14:16]
	second := afterDate[17:19]

	compareDate := year + "-" + month + "-" + day + " " + hour + ":" + minute + ":" + second

	rows, err := db.Query("SELECT id, title, category, content, Created, updated FROM pages WHERE substr(updated, 7, 4) || '-' || substr(updated, 4, 2) || '-' || substr(updated, 1, 2) || ' ' || substr(updated, 12) > ?;", compareDate)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var id int
	var title string
	var category string
	var content string
	var created string
	var updated string

	for rows.Next() {
		err := rows.Scan(&id, &title, &category, &content, &created, &updated)
		if err != nil {
			return nil, err
		}
		var page model.Page
		page.Id = id
		page.Title = title
		page.Category = category
		page.Content = content
		page.Created = created
		page.Updated = updated
		pages = append(pages, page)
	}
	return pages, nil
}

func GetPagesFromCategoryWithoutContent(whereCategory string) ([]model.Page, error) {
	var pages []model.Page

	rows, err := db.Query("select id, title, created, updated from pages where category=?  order by title COLLATE NOCASE ASC", whereCategory)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var id int
	var title string
	var created string
	var updated string

	for rows.Next() {
		err := rows.Scan(&id, &title, &created, &updated)
		if err != nil {
			return nil, err
		}
		var page model.Page
		page.Id = id
		page.Title = title
		page.Content = ""
		page.Created = created
		page.Updated = updated
		pages = append(pages, page)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return pages, nil
}

func GetPage(idGiven int) (model.Page, error) {
	var page model.Page

	rows, err := db.Query("select id, category, title, content, created, updated from pages where id=?", idGiven)
	if err != nil {
		return page, err
	}

	defer rows.Close()

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

func GetPages(ids []int) ([]model.Page, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf("select id, category, title, content, created, updated from pages where id IN (%s)", strings.Join(placeholders, ","))
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages []model.Page
	for rows.Next() {
		var p model.Page
		err := rows.Scan(&p.Id, &p.Category, &p.Title, &p.Content, &p.Created, &p.Updated)
		if err != nil {
			return nil, err
		}
		pages = append(pages, p)
	}
	return pages, nil
}

func GetPageFromOffset(recordTh int) (model.Page, error) {
	var page model.Page

	rows, err := db.Query("SELECT id, content FROM pages LIMIT 1 OFFSET ?;", recordTh)
	if err != nil {
		return page, err
	}

	defer rows.Close()

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

	rows, err := db.Query("SELECT id, category FROM pages where LOWER(title)=?;", titleGiven)
	if err != nil {
		return false
	}

	defer rows.Close()

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

	rows, err := db.Query("select image from pictures where id=?", id)
	if err != nil {
		return nil
	}

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

func GetImagesFromDateAfter(afterDate string) ([]model.Picture, error) {
	var images []model.Picture

	year := afterDate[6:10]
	month := afterDate[3:5]
	day := afterDate[0:2]
	hour := afterDate[11:13]
	minute := afterDate[14:16]
	second := afterDate[17:19]

	compareDate := year + "-" + month + "-" + day + " " + hour + ":" + minute + ":" + second

	rows, err := db.Query("SELECT * FROM pictures WHERE substr(created, 7, 4) || '-' || substr(created, 4, 2) || '-' || substr(created, 1, 2) || ' ' || substr(created, 12) > ?;", compareDate)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var id string
	var image []byte
	var created string
	var updated string

	for rows.Next() {
		err := rows.Scan(&id, &image, &created, &updated)
		if err != nil {
			return nil, err
		}
		var img model.Picture
		img.Id = id
		img.ImageBytes = image
		img.Created = created
		img.Updated = updated
		images = append(images, img)
	}
	return images, nil
}

func GetImagesFromDateAfterUpdated(afterDate string) ([]model.Picture, error) {
	var images []model.Picture

	year := afterDate[6:10]
	month := afterDate[3:5]
	day := afterDate[0:2]
	hour := afterDate[11:13]
	minute := afterDate[14:16]
	second := afterDate[17:19]

	compareDate := year + "-" + month + "-" + day + " " + hour + ":" + minute + ":" + second

	rows, err := db.Query("SELECT * FROM pictures WHERE substr(updated, 7, 4) || '-' || substr(updated, 4, 2) || '-' || substr(updated, 1, 2) || ' ' || substr(updated, 12) > ?;", compareDate)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var id string
	var image []byte
	var created string
	var updated string

	for rows.Next() {
		err := rows.Scan(&id, &image, &created, &updated)
		if err != nil {
			return nil, err
		}
		var img model.Picture
		img.Id = id
		img.ImageBytes = image
		img.Created = created
		img.Updated = updated
		images = append(images, img)
	}
	return images, nil
}

func Get25biggestImages() (error, []model.PictureInfo) {

	var response []model.PictureInfo

	rows, err := db.Query("SELECT id, image_size_bytes, image FROM pictures WHERE image_size_bytes IS NOT NULL AND id NOT LIKE '%.gif' ORDER BY image_size_bytes DESC LIMIT 25;")
	if err != nil {
		return err, response
	}

	defer rows.Close()

	var id string
	var image_size_bytes int
	var image []byte

	for rows.Next() {
		err := rows.Scan(&id, &image_size_bytes, &image)
		if err != nil {
			return nil, response
		}
		response = append(response, model.PictureInfo{Id: id, ImageSizeBytes: image_size_bytes, Image: image})
	}
	return nil, response
}

func GetImages() (error, []model.PictureInfo) {

	var response []model.PictureInfo

	rows, err := db.Query("SELECT id, image_size_bytes FROM pictures WHERE image_size_bytes IS NOT NULL;")
	if err != nil {
		return err, response
	}

	defer rows.Close()

	var id string
	var image_size_bytes int

	for rows.Next() {
		err := rows.Scan(&id, &image_size_bytes)
		if err != nil {
			return nil, response
		}
		response = append(response, model.PictureInfo{Id: id, ImageSizeBytes: image_size_bytes})
	}
	return nil, response
}

func UpdatePage(newPage model.Page) error {
	stmt, err := db.Prepare("UPDATE pages SET category=?, title=?, content=?, Updated=? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	newPage.Category = strings.ToLower(newPage.Category)

	res, err := stmt.Exec(newPage.Category, newPage.Title, newPage.Content, helpers.GetCurrentDateTime(), newPage.Id)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if affected == 1 {
		return nil
	}
	return errors.New("nothing updated")
}

func AddPage(page model.Page) (int, error) {
	var id int
	var rows *sql.Rows

	rows, err := db.Query("select id from pages where title=?", page.Title)
	if err != nil {
		return 0, err
	}
	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			return 0, err
		}

		if id > 0 {
			rows.Close()
			return 0, errors.New("already exist.")
		}
	}
	rows.Close()

	stmt, err := db.Prepare("INSERT INTO pages (title, category, content, created, updated) VALUES (?,?,?,?,?) RETURNING id")
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
	rows, err := db.Query("SELECT COUNT() as count FROM Pages")
	if err != nil {
		fmt.Printf("%s", err)
		return 0, err
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

	fmt.Println("select id, category, title, content, created, updated from pages where" + queryPart)
	rows, err := db.Query("select id, category, title, content, created, updated from pages where" + queryPart)
	if err != nil {
		return pages, err
	}

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

func GetPagesFromDates(dates ...string) ([]model.Page, error) {
	var pages []model.Page

	if len(dates) == 0 {
		return pages, nil
	}

	placeholders := make([]string, len(dates))
	args := make([]any, len(dates))
	for i, d := range dates {
		placeholders[i] = "?"
		args[i] = d
	}

	query := fmt.Sprintf(
		`SELECT id, title, created, category
     FROM pages
     WHERE substr(created, 1, 10) IN (%s)
     ORDER BY title COLLATE NOCASE ASC`,
		strings.Join(placeholders, ","),
	)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var p model.Page
		if err := rows.Scan(&p.Id, &p.Title, &p.Created, &p.Category); err != nil {
			return nil, err
		}
		pages = append(pages, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return pages, nil
}

func AddImage(id string, data []byte) bool {
	stmt, err := db.Prepare("INSERT INTO pictures (id, image, created, updated) VALUES (?,?,?,?)")
	if err != nil {
		log.Println(err)
		return false
	}
	defer stmt.Close()

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
	stmt, err := db.Prepare("delete from pictures where id=?")
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

	stmt, err := db.Prepare("delete from pages where id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()
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

func AddPrePage(url string) error {
	var id int
	var rows *sql.Rows

	rows, err := db.Query("select id from prepages where url=?", url)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			return err
		}

		if id > 0 {
			return errors.New("already exist.")
		}
	}

	stmt, err := db.Prepare("INSERT INTO prepages (url, created) VALUES (?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(url, helpers.GetCurrentDateTime())
	if err != nil {
		return err
	}

	return nil
}

func GetAllPrePages() ([]model.PrePage, error) {
	var pages []model.PrePage

	rows, err := db.Query("select id, url, created from prepages order by created COLLATE NOCASE ASC")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var pp model.PrePage
		if err := rows.Scan(&pp.Id, &pp.Url, &pp.Created); err != nil {
			return nil, err
		}
		pages = append(pages, pp)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return pages, nil
}

func DeletePrePage(id int) error {

	stmt, err := db.Prepare("delete from prepages where id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	res, err := stmt.Exec(id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 1 {
		return nil
	}

	return errors.New("id doesn't exist")
}

func GetLinkCategories() ([]string, error) {
	var categories []string
	rows, err := db.Query("select DISTINCT category from links order by category")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

	rows, err := db.Query("select id, url, description, category, created, IFNULL(updated, '') from links where category=?  order by description COLLATE NOCASE ASC", whereCategory)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

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
	rows, err := db.Query("select category, url, description, created, IFNULL(updated, '') from links where id=?", idGiven)
	if err != nil {
		return link, err
	}
	defer rows.Close()

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

	rows, err := db.Query("select id from links where url = ?", givenUrl)
	if err != nil {
		return false
	}

	defer rows.Close()

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
	stmt, err := db.Prepare("UPDATE links SET category=?, url=?, description=?, Updated=? WHERE id = ?")
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
	return errors.New("nothing updated")
}

func AddLink(newLink model.Link) (int, error) {
	var id int
	var rows *sql.Rows

	rows, err := db.Query("select id from links where url=?", newLink.Url)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			return 0, err
		}

		if id > 0 {
			return 0, errors.New("already exist")
		}
	}

	stmt, err := db.Prepare("INSERT INTO links(category, url, description, created) VALUES (?,?,?,?) RETURNING id")
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
	stmt, err := db.Prepare("delete from links where id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()
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

	abbrMu.Lock()
	defer abbrMu.Unlock()

	rows, err := db.Query("select id, name, description from abbreviations")
	if err != nil {
		return
	}
	defer rows.Close()

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
		if len(abbreviation.Name) == 0 {
			continue
		}
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
	abbrMu.RLock()
	defer abbrMu.RUnlock()
	return abbreviationsInCache[givenLetter], nil
}

func GetAllAbbreviations() ([]model.Abbreviation, error) {
	var abbreviation model.Abbreviation
	var abbreviations []model.Abbreviation
	rows, err := db.Query("select id, name, description from abbreviations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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
		abbreviations = append(abbreviations, abbreviation)
	}
	return abbreviations, nil
}

func AddAbbreviation(abbreviation model.Abbreviation) (int, error) {
	var id int
	stmt, err := db.Prepare("INSERT INTO abbreviations (name, description) VALUES (?,?) RETURNING id")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	err = stmt.QueryRow(strings.ToUpper(abbreviation.Name), abbreviation.Description).Scan(&id)
	if err != nil {
		return 0, err
	}

	abbrMu.Lock()
	defer abbrMu.Unlock()
	abbrss := abbreviationsInCache[strings.ToUpper(abbreviation.Name[0:1])]
	abbrss = append(abbrss, abbreviation)
	abbreviationsInCache[strings.ToUpper(abbreviation.Name[0:1])] = abbrss
	return id, nil
}

func DeleteAbbreviation(id string) error {
	stmt, err := db.Prepare("Delete from abbreviations where id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}

	setAbbreviationInCache()
	return nil
}

func deleteElement(slice []model.Abbreviation, index int) []model.Abbreviation {
	return append(slice[:index], slice[index+1:]...)
}

func GetImageCount() (int, error) {
	rows, err := db.Query("SELECT COUNT() as count FROM pictures")
	if err != nil {
		fmt.Printf("%s", err)
		return 0, err
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

	rows, err := db.Query("SELECT id FROM pictures LIMIT 1 OFFSET ?;", recordTh)
	if err != nil {
		return image, err
	}

	defer rows.Close()

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

func Getstatistics() model.Statistic {
	var rtn model.Statistic

	count, err := GetPagesCount()
	if err != nil {
		return rtn
	}
	rtn.Total = count

	rows, err := db.Query("SELECT category, COUNT(*) as count FROM pages GROUP BY category ORDER BY count DESC, category ASC")
	if err != nil {
		log.Println("Error getting statistics:", err)
		return rtn
	}
	defer rows.Close()

	for rows.Next() {
		var c string
		var i int
		if err := rows.Scan(&c, &i); err != nil {
			log.Println("Error scanning statistics:", err)
			continue
		}
		cat := model.SCategory{Name: c, Count: i}
		rtn.SCategories = append(rtn.SCategories, cat)
	}

	return rtn
}

func Play() {
	stmt, err := db.Prepare("CREATE INDEX IF NOT EXISTS idx_pictures_size_desc ON pictures(image_size_bytes DESC, id);")
	if err != nil {
		return
	}

	stmt.Exec()

	defer stmt.Close()

	fmt.Println("Prima de luxe")
}
