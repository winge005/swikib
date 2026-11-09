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

// RunAnalysis analyses all local pages for duplicates. When resume is nil a new
// analysis is started. The returned bool is true when the analysis stopped
// because a pause was requested; in that case the checkpoint has been saved.
func RunAnalysis(resume *AnalysisState) (bool, error) {
	analysisMu.Lock()
	defer analysisMu.Unlock()

	if db == nil {
		if err := InitDB(); err != nil {
			return false, err
		}
	}

	categories, err := GetCategories()
	if err != nil {
		return false, err
	}

	totalPages, err := GetPagesCount()
	if err != nil {
		totalPages = 0
	}

	state := &AnalysisState{Pairs: []SimilarPagePair{}}
	startCategory := 0
	startPageIndex := 0
	startPageID := 0
	completedPages := 0
	initialPercent := 0

	if resume != nil {
		state = resume
		if state.Pairs == nil {
			state.Pairs = []SimilarPagePair{}
		}
		startPageIndex = state.PageIndex
		startPageID = state.LastPageID
		completedPages = state.CompletedPages
		initialPercent = state.Percent
		startCategory = resolveCategoryIndex(categories, state.Category, state.CategoryIndex)
		state.Paused = false
		state.Completed = false
	}

	if totalPages <= 0 {
		totalPages = completedPages + 1
	}

	setProgress(initialPercent, "Analyseren...")

	for ci := startCategory; ci < len(categories); ci++ {
		category := categories[ci]

		pageList, err := GetPagesFromCategoryWithContent(category)
		if err != nil {
			return false, err
		}
		if len(pageList) < 2 {
			completedPages += len(pageList)
			continue
		}

		docs := make([]analysisDoc, len(pageList))
		for i, page := range pageList {
			if PauseRequested() {
				nextIndex := 0
				lastID := 0
				if ci == startCategory && resume != nil {
					nextIndex = startPageIndex
					lastID = startPageID
				}
				if err := savePausedState(state, category, ci, nextIndex, lastID, completedPages, totalPages); err != nil {
					return false, err
				}
				return true, nil
			}

			docs[i] = analysisDoc{
				page:     page,
				tokens:   similarity.Tokenize(page.Title),
				shingles: similarity.Shingles(page.Content, 3),
			}

			prepared := completedPages + i + 1
			setProgress(stepPercent(0, 40, prepared, totalPages), fmt.Sprintf("Voorbereiden: %d/%d pagina's", prepared, totalPages))
		}

		index := make(map[string][]int)
		for i := range docs {
			for shingle := range docs[i].shingles {
				index[shingle] = append(index[shingle], i)
			}
		}

		analyzeFrom := 0
		if ci == startCategory && resume != nil {
			analyzeFrom = resolveStartIndex(docs, startPageIndex, startPageID)
		}

		seen := make(map[[2]int]bool)
		for _, pair := range state.Pairs {
			if pair.Category == category {
				seen[[2]int{pair.Page1Id, pair.Page2Id}] = true
			}
		}

		for i := analyzeFrom; i < len(docs); i++ {
			if PauseRequested() {
				lastID := 0
				if i > 0 {
					lastID = docs[i-1].page.Id
				}
				if err := savePausedState(state, category, ci, i, lastID, completedPages, totalPages); err != nil {
					return false, err
				}
				return true, nil
			}

			analyzed := completedPages + i + 1
			setProgress(stepPercent(40, 60, analyzed, totalPages), fmt.Sprintf("Analyseren: %d/%d pagina's", analyzed, totalPages))

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

				state.Pairs = append(state.Pairs, SimilarPagePair{
					Score:      score,
					Category:   category,
					Page1Id:    page1.Id,
					Page1Title: page1.Title,
					Page2Id:    page2.Id,
					Page2Title: page2.Title,
				})
			}
		}

		completedPages += len(docs)

		state.Category = ""
		if ci+1 < len(categories) {
			state.Category = categories[ci+1]
		}
		state.CategoryIndex = ci + 1
		state.PageIndex = 0
		state.LastPageID = 0
		state.CompletedPages = completedPages
		state.TotalPages = totalPages
		state.Percent = CurrentProgress().Percent
		state.Paused = false
		state.Completed = false
		sortSimilarPagePairs(state.Pairs)
		if err := SaveAnalysisState(state); err != nil {
			return false, err
		}
	}

	sortSimilarPagePairs(state.Pairs)
	state.Completed = true
	state.Paused = false
	state.Percent = 100
	state.Category = ""
	state.CategoryIndex = len(categories)
	state.PageIndex = 0
	state.LastPageID = 0
	state.CompletedPages = completedPages
	state.TotalPages = totalPages
	if err := SaveAnalysisState(state); err != nil {
		return false, err
	}

	setProgress(100, fmt.Sprintf("%d duplicaten gevonden", len(state.Pairs)))

	return false, nil
}

func savePausedState(state *AnalysisState, category string, categoryIndex, pageIndex, lastPageID, completedPages, totalPages int) error {
	state.Category = category
	state.CategoryIndex = categoryIndex
	state.PageIndex = pageIndex
	state.LastPageID = lastPageID
	state.CompletedPages = completedPages
	state.TotalPages = totalPages
	state.Percent = CurrentProgress().Percent
	state.Paused = true
	state.Completed = false
	sortSimilarPagePairs(state.Pairs)

	return SaveAnalysisState(state)
}

func resolveCategoryIndex(categories []string, category string, index int) int {
	if category != "" {
		for i, available := range categories {
			if available == category {
				return i
			}
		}
	}

	if index < 0 {
		return 0
	}
	if index > len(categories) {
		return len(categories)
	}

	return index
}

func resolveStartIndex(docs []analysisDoc, pageIndex, lastPageID int) int {
	if lastPageID > 0 {
		for i := range docs {
			if docs[i].page.Id == lastPageID {
				return i + 1
			}
		}
	}

	if pageIndex < 0 {
		return 0
	}
	if pageIndex > len(docs) {
		return len(docs)
	}

	return pageIndex
}

func sortSimilarPagePairs(pairs []SimilarPagePair) {
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
}
