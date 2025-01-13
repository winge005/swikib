package images

import (
	"fmt"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"regexp"
	"strings"
	"swiki/persistence"
	"time"
)

func CheckUneededImages() {

	currentTime := time.Now()

	nr, error := persistence.GetPagesCount()
	if error != nil {
		fmt.Println(error.Error())
		return
	}

	var currentNr = 0
	var imagesUsedInPages []string
	var imagesInDb []string
	var totalDeleted = 0

	fmt.Println(time.Since(currentTime), "nr of pages: ", nr)

	for currentNr < nr {
		time.Sleep(time.Millisecond * 5)
		page, err := persistence.GetPageFromOffset(currentNr)
		if err != nil {
			fmt.Println(err.Error())
		}

		md := []byte(page.Content)
		htmlFlags := html.CommonFlags | html.HrefTargetBlank
		opts := html.RendererOptions{Flags: htmlFlags}
		renderer := html.NewRenderer(opts)
		page.Content = string(markdown.ToHTML(md, nil, renderer))
		if page.Id == 1296 {
			fmt.Println("")
		}
		imagesFound := ProcessImagesFromHtml(page.Content)
		if len(imagesFound) > 0 {
			for _, v := range imagesFound {
				imagesUsedInPages = append(imagesUsedInPages, v)
			}
		}

		currentNr += 1
	}

	fmt.Println(time.Since(currentTime), "pages in slice")

	nr, error = persistence.GetImageCount()
	if error != nil {
		fmt.Println(error.Error())
		return
	}
	fmt.Println(time.Since(currentTime), "nr of pictures", nr)
	currentNr = 0
	for currentNr < nr {
		time.Sleep(time.Millisecond * 2)
		image, err := persistence.GetImageFromOffSet(currentNr)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		imagesInDb = append(imagesInDb, image)
		currentNr += 1
	}

	fmt.Println(time.Since(currentTime), "pictures in slice")

	var imagesToDelete []string

	for _, v := range imagesInDb {
		if !contains(imagesUsedInPages, v) {
			imagesToDelete = append(imagesToDelete, v)
		}
	}

	fmt.Println(time.Since(currentTime), "pictures to delete", len(imagesToDelete))

	for _, v := range imagesToDelete {
		result, err := persistence.DeleteImage(v)
		if err != nil {
			fmt.Println(err.Error())
		}

		totalDeleted += int(result)
	}

	fmt.Println(time.Since(currentTime), "ready with deleting ", totalDeleted)
}

func contains(elems []string, v string) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}

func ProcessImagesFromHtml(html string) []string {
	repsonse := make([]string, 0)
	r := regexp.MustCompile(`<img[^>]*src="([^"]+)"[^>]*>`)

	matches := r.FindAllString(html, -1)
	var idx = 0
	for _, v := range matches {
		bPos := strings.Index(v[idx:], "src=\"") + idx

		if bPos > -1 {
			ePos := strings.Index(v[bPos+len("src=\""):], "\"") + len("src=\"")
			repsonse = append(repsonse, v[bPos+5:ePos+5])
		}
	}

	return repsonse
}
