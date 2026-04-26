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
	lastSync, err := os.ReadFile("lsync.txt")
	if err != nil {
		panic(err)
	}

	updatePages(string(lastSync))
	updateImages(string(lastSync))
	updateAbbreviations(string(lastSync))
	updateLinks(string(lastSync))
	elapsed := time.Since(start)
	log.Printf("Sync complete total: %v", elapsed)
}

func Sync() {

	start := time.Now()
	clearDB()
	CreateTables()

	addPages()
	addImages()
	addLinks()
	addAbbreviations()
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

	for _, v := range categories {
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
			log.Printf("Added page %v", i)
		}
	}
	log.Println("pages added")
}

func addImages() {
	err, pictureInfos := persistence.GetImages()
	if err != nil {
		println(err)
		return
	}

	for _, v := range pictureInfos {
		pil := model.Picture{ImageBytes: v.Image, Id: v.Id}

		result := AddImage(pil)
		if !result {
			log.Printf("Error adding image %v", v.Id)
			return
		}

		log.Printf("added image %v", pil.Id)
	}
	log.Println("Images added")
}

func addLinks() {
	categories, err := persistence.GetLinkCategories()
	if err != nil {
		println(err)
		return
	}

	for _, v := range categories {
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
	}
	log.Println("links added")
}

func addAbbreviations() {
	abbreviations, err := persistence.GetAllAbbreviations()
	if err != nil {
		println(err)
		return
	}

	for _, abr := range abbreviations {
		abrLocal := model.AbbreviationLocal{Name: abr.Name, Description: abr.Description, TursoId: abr.Id}
		i, err := AddAbbreviation(abrLocal)
		if err != nil {
			log.Printf("Error adding abbreviation %v: %v", i, err)
			continue
		}
		log.Printf("Added abbreviation %v, %v", abrLocal.TursoId, abr.Name)
	}
	log.Println("abbreviations added")
}

// updated in this context means pages that have been added after last sync. Also all pages that have bene updated after last sync.
func updatePages(lastSync string) {
	pages, err := persistence.GetPagesFromDateAndAfterWithContent(lastSync)
	if err != nil {
		return
	}
	for _, p := range pages {
		pl := model.PageLocal{Category: p.Category, Title: p.Title, Content: p.Content, Created: p.Created, Updated: p.Updated, TursoId: p.Id}
		i, err := AddPage(pl)
		if err != nil {
			log.Printf("Error adding page %v: %v", p, err)
			continue
		}
		log.Printf("Added page %v", i)
	}

	pages, err = persistence.GetUpdatedPagesFromDateAndAfterWithContent(lastSync)
	if err != nil {
		return
	}
	for _, p := range pages {
		pl := model.PageLocal{Category: p.Category, Title: p.Title, Content: p.Content, Created: p.Created, Updated: p.Updated, TursoId: p.Id}
		err := UpdatePage(pl)
		if err != nil {
			log.Printf("Error updating page %v: %v", p, err)
			continue
		}
		log.Printf("Added page %v", pl.TursoId)
	}
}

func updateImages(lastSync string) {
	images, err := persistence.GetImagesFromDateAfter(lastSync)
	if err != nil {
		return
	}
	for _, img := range images {
		imgLocal := model.Picture{Id: img.Id, ImageBytes: img.ImageBytes, Created: img.Created, Updated: img.Updated}
		succedeed := AddImage(imgLocal)
		if !succedeed {
			log.Printf("Error adding image %v", img.Id)
			continue
		}
		log.Printf("Added image %v", img.Id)
	}
}

func updateAbbreviations(lastSync string) {
	abbreviations, err := persistence.GetAllAbbreviations()
	if err != nil {
		return
	}
	for _, abr := range abbreviations {
		abbreviation, err := GetAbbreviation(abr.Name)
		if err != nil {
			log.Printf("Error reading abbriviation from local %v", abr.Name)
			continue
		}

		abbreviationLocal := model.AbbreviationLocal{Name: abbreviation.Name, Description: abbreviation.Description, TursoId: abbreviation.Id}
		i, err := AddAbbreviation(abbreviationLocal)
		if err != nil {
			log.Printf("Error adding abbreviation to local %v", abbreviationLocal.Name)
			continue
		}
		log.Printf("added abbreviation to local id: %v", i)
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
