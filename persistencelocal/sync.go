package persistencelocal

import (
	"fmt"
	"log"
	"swiki/model"
	"swiki/persistence"
	"time"
)

func Sync() {
	start := time.Now()

	addPages()
	addImages()
	addLinks()
	addAbbreviations()
	elapsed := time.Since(start)
	log.Printf("Sync complete total: %v", elapsed)
}

func addPages() {
	categories, err := persistence.GetCategories()
	if err != nil {
		println(err)
		return
	}

	start := time.Now()
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

	elapsed := time.Since(start)
	log.Printf("Sync complete pages: %v", elapsed)
}

func addImages() {
	err, pictureInfos := persistence.GetImages()
	if err != nil {
		println(err)
		return
	}

	start := time.Now()
	for _, v := range pictureInfos {
		pil := model.PictureInfoLocal{ImageSizeBytes: v.ImageSizeBytes, TursoId: v.Id}

		result := AddImage(pil)
		if !result {
			log.Printf("Error adding image %v", v.Id)
			return
		}

		log.Printf("added image %v", pil.TursoId)
	}

	elapsed := time.Since(start)
	log.Printf("Sync complete images: %v", elapsed)
}

func addLinks() {
	categories, err := persistence.GetLinkCategories()
	if err != nil {
		println(err)
		return
	}

	start := time.Now()
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

	elapsed := time.Since(start)
	log.Printf("Sync complete links: %v", elapsed)
}

func addAbbreviations() {
	start := time.Now()
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

	elapsed := time.Since(start)
	log.Printf("Sync complete abbreviations: %v", elapsed)
}
