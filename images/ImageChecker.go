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
	nr, error := persistence.GetPagesCount()
	if error != nil {
		fmt.Println(error.Error())
		return
	}

	var currentNr = 0

	for currentNr < nr {
		time.Sleep(time.Millisecond * 10)
		page, err := persistence.GetPageFromOfset(currentNr)
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
				fmt.Printf("%v: %v\n", page.Id, v)
			}
		}

		//fmt.Printf("id: %v\n", page.Id)
		currentNr += 1
	}

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
