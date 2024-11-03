package helpers

import (
	"regexp"
	"strings"
)

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
