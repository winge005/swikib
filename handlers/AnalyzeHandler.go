package handlers

import (
	"fmt"
	"net/http"
	"swiki/model"
	"swiki/persistencelocal"

	"github.com/agnivade/levenshtein"
)

type SimilarPair struct {
	Page1    model.PageLocal
	Page2    model.PageLocal
	Distance int
}

func AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	//var resultSet []SimilarPair
	categories, err := persistencelocal.GetCategories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for _, v := range categories {
		pages, err := persistencelocal.GetPagesFromCategoryWithContent(v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		pairs := findSimilarPages(pages, 3)
		for _, p := range pairs {
			fmt.Printf(
				"%q <-> %q | distance = %d\n",
				p.Page1.Title,
				p.Page2.Title,
				p.Distance,
			)
		}
	}
}

func findSimilarPages(pages []model.PageLocal, maxDistance int) []SimilarPair {
	var result []SimilarPair

	for i := 0; i < len(pages); i++ {
		for j := i + 1; j < len(pages); j++ {
			d := levenshtein.ComputeDistance(pages[i].Content, pages[j].Content)
			fmt.Printf("%v\n", d)
			if d <= maxDistance {
				result = append(result, SimilarPair{
					Page1:    pages[i],
					Page2:    pages[j],
					Distance: d,
				})
			}
		}
	}

	return result
}
