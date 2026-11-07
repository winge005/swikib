package persistencelocal

import (
	"fmt"
	"log"
	"os"
	"swiki/helpers"
	"swiki/model"
	"swiki/persistence"
	"time"
)

func UpdateSync() {
	start := time.Now()
	setProgress(1, "Laatste sync-tijd lezen...")
	lastSync, err := os.ReadFile("lsync.txt")
	if err != nil {
		panic(err)
	}

	updatePages(string(lastSync))
	updateImages(string(lastSync))
	updateAbbreviations(string(lastSync))
	updateLinks(string(lastSync))
	setProgress(100, "Update voltooid")
	elapsed := time.Since(start)
	log.Printf("Sync complete total: %v", elapsed)
}

func Sync() {

	start := time.Now()
	setProgress(1, "Lokale database leegmaken...")
	clearDB()
	CreateTables()
	setProgress(3, "Tabellen aangemaakt")

	addPages()
	addImages()
	addLinks()
	addAbbreviations()
	setProgress(100, "Volledige sync voltooid")
	elapsed := time.Since(start)
	log.Printf("Sync complete total: %v", elapsed)

	data := []byte(helpers.GetCurrentDateTime())

	err := os.WriteFile("lsync.txt", data, 0644)
	if err != nil {
		panic(err)
	}
}

func addPages() {
	categories, err := persistence.GetCategories()
	if err != nil {
		println(err)
		return
	}

	total, err := persistence.GetPagesCount()
	if err != nil || total <= 0 {
		total = 0
	}

	processed := 0
	for ci, v := range categories {
		fmt.Println(v)
		pages, err := persistence.GetPagesFromCategoryWithContent(v)
		if err != nil {
			println(err)
			return
		}

		for _, p := range pages {
			pl := model.PageLocal{Category: p.Category, Title: p.Title, Content: p.Content, Created: p.Created, Updated: p.Updated, TursoId: p.Id}
			i, err := AddPage(pl)
			if err != nil {
				log.Printf("Error adding page %v: %v", p, err)
				continue
			}
			processed++
			if total > 0 {
				setProgress(stepPercent(3, 62, processed, total), fmt.Sprintf("Pagina's: %d/%d", processed, total))
			} else {
				setProgress(stepPercent(3, 62, ci+1, len(categories)), fmt.Sprintf("Categorieën: %d/%d", ci+1, len(categories)))
			}
			log.Printf("Added page %v", i)
		}
	}
	setProgress(65, "Pagina's toegevoegd")
	log.Println("pages added")
}

func addImages() {
	err, pictureInfos := persistence.GetImages()
	if err != nil {
		println(err)
		return
	}

	total := len(pictureInfos)
	for i, v := range pictureInfos {
		pil := model.Picture{ImageBytes: v.Image, Id: v.Id}

		result := AddImage(pil)
		if !result {
			log.Printf("Error adding image %v", v.Id)
			return
		}

		setProgress(stepPercent(65, 20, i+1, total), fmt.Sprintf("Afbeeldingen: %d/%d", i+1, total))
		log.Printf("added image %v", pil.Id)
	}
	setProgress(85, "Afbeeldingen toegevoegd")
	log.Println("Images added")
}

func addLinks() {
	categories, err := persistence.GetLinkCategories()
	if err != nil {
		println(err)
		return
	}

	for ci, v := range categories {
		fmt.Println(v)
		links, err := persistence.GetLinksFromCategory(v)
		if err != nil {
			println(err)
			return
		}

		for _, l := range links {
			nl := model.LinkLocal{Category: l.Category, Url: l.Url, Description: l.Description, Created: l.Created, Updated: l.Updated, TursoId: l.Id}
			i, err := AddLink(nl)
			if err != nil {
				log.Printf("Error adding link %v: %v", l.Id, err)
				continue
			}
			log.Printf("Added link %v", i)
		}

		setProgress(stepPercent(85, 10, ci+1, len(categories)), fmt.Sprintf("Links: categorie %d/%d", ci+1, len(categories)))
	}
	setProgress(95, "Links toegevoegd")
	log.Println("links added")
}

func addAbbreviations() {
	abbreviations, err := persistence.GetAllAbbreviations()
	if err != nil {
		println(err)
		return
	}

	total := len(abbreviations)
	for i, abr := range abbreviations {
		abrLocal := model.AbbreviationLocal{Name: abr.Name, Description: abr.Description, TursoId: abr.Id}
		added, err := AddAbbreviation(abrLocal)
		if err != nil {
			log.Printf("Error adding abbreviation %v: %v", added, err)
			continue
		}
		setProgress(stepPercent(95, 5, i+1, total), fmt.Sprintf("Afkortingen: %d/%d", i+1, total))
		log.Printf("Added abbreviation %v, %v", abrLocal.TursoId, abr.Name)
	}
	setProgress(100, "Afkortingen toegevoegd")
	log.Println("abbreviations added")
}

// updated in this context means pages that have been added after last sync. Also all pages that have bene updated after last sync.
func updatePages(lastSync string) {
	setProgress(5, "Nieuwe pagina's ophalen...")
	pages, err := persistence.GetPagesFromDateAndAfterWithContent(lastSync)
	if err != nil {
		return
	}
	total := len(pages)
	for i, p := range pages {
		pl := model.PageLocal{Category: p.Category, Title: p.Title, Content: p.Content, Created: p.Created, Updated: p.Updated, TursoId: p.Id}
		added, err := AddPage(pl)
		if err != nil {
			log.Printf("Error adding page %v: %v", p, err)
			continue
		}
		setProgress(stepPercent(5, 35, i+1, total), fmt.Sprintf("Nieuwe pagina's: %d/%d", i+1, total))
		log.Printf("Added page %v", added)
	}

	setProgress(40, "Gewijzigde pagina's ophalen...")
	pages, err = persistence.GetUpdatedPagesFromDateAndAfterWithContent(lastSync)
	if err != nil {
		return
	}
	total = len(pages)
	for i, p := range pages {
		pl := model.PageLocal{Category: p.Category, Title: p.Title, Content: p.Content, Created: p.Created, Updated: p.Updated, TursoId: p.Id}
		err := UpdatePage(pl)
		if err != nil {
			log.Printf("Error updating page %v: %v", p, err)
			continue
		}
		setProgress(stepPercent(40, 30, i+1, total), fmt.Sprintf("Gewijzigde pagina's: %d/%d", i+1, total))
		log.Printf("Updated page %v", pl.TursoId)
	}
}

func updateImages(lastSync string) {
	setProgress(70, "Afbeeldingen ophalen...")
	images, err := persistence.GetImagesFromDateAfter(lastSync)
	if err != nil {
		return
	}
	total := len(images)
	for i, img := range images {
		imgLocal := model.Picture{Id: img.Id, ImageBytes: img.ImageBytes, Created: img.Created, Updated: img.Updated}
		succedeed := AddImage(imgLocal)
		if !succedeed {
			log.Printf("Error adding image %v", img.Id)
			continue
		}
		setProgress(stepPercent(70, 20, i+1, total), fmt.Sprintf("Afbeeldingen: %d/%d", i+1, total))
		log.Printf("Added image %v", img.Id)
	}
}

func updateAbbreviations(lastSync string) {
	setProgress(90, "Afkortingen ophalen...")
	abbreviations, err := persistence.GetAllAbbreviations()
	if err != nil {
		return
	}
	total := len(abbreviations)
	for i, abr := range abbreviations {
		abbreviation, err := GetAbbreviation(abr.Name)
		if err != nil {
			log.Printf("Error reading abbriviation from local %v", abr.Name)
			continue
		}

		abbreviationLocal := model.AbbreviationLocal{Name: abbreviation.Name, Description: abbreviation.Description, TursoId: abbreviation.Id}
		added, err := AddAbbreviation(abbreviationLocal)
		if err != nil {
			log.Printf("Error adding abbreviation to local %v", abbreviationLocal.Name)
			continue
		}
		setProgress(stepPercent(90, 10, i+1, total), fmt.Sprintf("Afkortingen: %d/%d", i+1, total))
		log.Printf("added abbreviation to local id: %v", added)
	}
}

func updateLinks(lastSync string) {

}

func clearDB() {
	err := DropTable("pages")
	if err != nil {
		return
	}
	err = DropTable("abbreviations")
	if err != nil {
		return
	}
	err = DropTable("links")
	if err != nil {
		return
	}
	err = DropTable("pages")
	if err != nil {
		return
	}
	err = DropTable("pictures")
	if err != nil {
		return
	}
}
