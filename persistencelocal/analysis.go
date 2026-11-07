package persistencelocal

import (
	"fmt"
	"sort"
	"sync"

	"swiki/model"
	"swiki/similarity"
)

const (
	analysisCandidateLimit = 20
	analysisMaxPostings    = 64
	analysisThreshold      = 0.45
)

var analysisMu sync.Mutex

type SimilarPagePair struct {
	Score      float64 `json:"score"`
	Category   string  `json:"category"`
	Page1Id    int     `json:"page1_id"`
	Page1Title string  `json:"page1_title"`
	Page2Id    int     `json:"page2_id"`
	Page2Title string  `json:"page2_title"`
}

type analysisDoc struct {
	page     model.PageLocal
	tokens   map[string]struct{}
	shingles map[string]struct{}
}

func FindSimilarPagePairs() ([]SimilarPagePair, error) {
	analysisMu.Lock()
	defer analysisMu.Unlock()

	if db == nil {
		if err := InitDB(); err != nil {
			return nil, err
		}
	}

	categories, err := GetCategories()
	if err != nil {
		return nil, err
	}

	pairs := make([]SimilarPagePair, 0)
	totalPages, _ := GetPagesCount()
	prepared := 0
	analyzed := 0

	for _, category := range categories {
		pageList, err := GetPagesFromCategoryWithContent(category)
		if err != nil {
			return nil, err
		}
		if len(pageList) < 2 {
			continue
		}

		docs := make([]analysisDoc, len(pageList))
		for i, page := range pageList {
			docs[i] = analysisDoc{
				page:     page,
				tokens:   similarity.Tokenize(page.Title),
				shingles: similarity.Shingles(page.Content, 3),
			}
			prepared++
			if totalPages > 0 {
				setProgress(stepPercent(0, 40, prepared, totalPages), fmt.Sprintf("Voorbereiden: %d/%d pagina's", prepared, totalPages))
			}
		}

		index := make(map[string][]int)
		for i := range docs {
			for shingle := range docs[i].shingles {
				index[shingle] = append(index[shingle], i)
			}
		}

		seen := make(map[[2]int]bool)

		for i := range docs {
			analyzed++
			if totalPages > 0 {
				setProgress(stepPercent(40, 60, analyzed, totalPages), fmt.Sprintf("Analyseren: %d/%d pagina's", analyzed, totalPages))
			}
			overlap := make(map[int]int)
			for shingle := range docs[i].shingles {
				postings := index[shingle]
				if len(postings) < 2 || len(postings) > analysisMaxPostings {
					continue
				}
				for _, j := range postings {
					if j != i {
						overlap[j]++
					}
				}
			}
			if len(overlap) == 0 {
				continue
			}

			candidates := make([]int, 0, len(overlap))
			for j := range overlap {
				candidates = append(candidates, j)
			}
			sort.Slice(candidates, func(a, b int) bool {
				if overlap[candidates[a]] != overlap[candidates[b]] {
					return overlap[candidates[a]] > overlap[candidates[b]]
				}
				return docs[candidates[a]].page.Id < docs[candidates[b]].page.Id
			})
			if len(candidates) > analysisCandidateLimit {
				candidates = candidates[:analysisCandidateLimit]
			}

			for _, j := range candidates {
				page1, page2 := docs[i].page, docs[j].page
				if page1.Id > page2.Id {
					page1, page2 = page2, page1
				}
				key := [2]int{page1.Id, page2.Id}
				if seen[key] {
					continue
				}
				seen[key] = true

				score := similarity.ScoreSets(docs[i].tokens, docs[i].shingles, docs[j].tokens, docs[j].shingles, true)
				if score < analysisThreshold {
					continue
				}

				pairs = append(pairs, SimilarPagePair{
					Score:      score,
					Category:   category,
					Page1Id:    page1.Id,
					Page1Title: page1.Title,
					Page2Id:    page2.Id,
					Page2Title: page2.Title,
				})
			}
		}
	}

	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].Score != pairs[j].Score {
			return pairs[i].Score > pairs[j].Score
		}
		if pairs[i].Category != pairs[j].Category {
			return pairs[i].Category < pairs[j].Category
		}
		if pairs[i].Page1Title != pairs[j].Page1Title {
			return pairs[i].Page1Title < pairs[j].Page1Title
		}
		return pairs[i].Page2Title < pairs[j].Page2Title
	})

	setProgress(100, fmt.Sprintf("%d duplicaten gevonden", len(pairs)))

	return pairs, nil
}
