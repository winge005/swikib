package search

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"swiki/helpers"
	"swiki/model"
	"swiki/persistence"
	"sync"
	"time"
)

var (
	idx   = newIndex()
	idxMu sync.RWMutex
)

var stopwords = map[string]struct{}{
	"a": {}, "about": {}, "above": {}, "after": {}, "again": {}, "against": {}, "all": {}, "am": {}, "an": {}, "and": {}, "any": {}, "are": {}, "aren't": {}, "as": {}, "at": {},
	"be": {}, "because": {}, "been": {}, "before": {}, "being": {}, "below": {}, "between": {}, "both": {}, "but": {}, "by": {},
	"can't": {}, "cannot": {}, "could": {}, "couldn't": {},
	"did": {}, "didn't": {}, "do": {}, "does": {}, "doesn't": {}, "doing": {}, "don't": {}, "down": {}, "during": {},
	"each": {},
	"few":  {}, "for": {}, "from": {}, "further": {},
	"had": {}, "hadn't": {}, "has": {}, "hasn't": {}, "have": {}, "haven't": {}, "having": {}, "he": {}, "he'd": {}, "he'll": {}, "he's": {}, "her": {}, "here": {}, "here's": {}, "hers": {}, "herself": {}, "him": {}, "himself": {}, "his": {}, "how": {}, "how's": {},
	"i": {}, "i'd": {}, "i'll": {}, "in": {}, "i'm": {}, "i've": {}, "if": {}, "into": {}, "is": {}, "isn't": {}, "it": {}, "it's": {}, "its": {}, "itself": {}, "let's": {},
	"me": {}, "more": {}, "most": {}, "mustn't": {}, "my": {}, "myself": {}, "no": {}, "nor": {}, "not": {},
	"of": {}, "off": {}, "on": {}, "once": {}, "only": {}, "or": {}, "other": {}, "ought": {}, "our": {}, "ours": {}, "ourselves": {}, "out": {}, "over": {}, "own": {},
	"same": {}, "shan't": {}, "she": {}, "she'd": {}, "she'll": {}, "she's": {}, "should": {}, "shouldn't": {}, "so": {}, "some": {}, "such": {}, "than": {},
	"that": {}, "that's": {}, "the": {}, "their": {}, "theirs": {}, "them": {}, "themselves": {}, "then": {}, "there": {}, "there's": {}, "these": {}, "they": {}, "they'd": {}, "they'll": {}, "they're": {}, "they've": {}, "this": {}, "those": {}, "through": {}, "to": {}, "too": {},
	"under": {}, "until": {}, "up": {},
	"very": {},
	"was":  {}, "wasn't": {}, "we": {}, "we'd": {}, "we'll": {}, "we're": {}, "we've": {}, "were": {}, "weren't": {}, "what": {}, "what's": {}, "when": {}, "when's": {}, "where": {}, "where's": {}, "which": {}, "while": {}, "who": {}, "who's": {}, "whom": {}, "why": {}, "why's": {}, "with": {}, "won't": {}, "would": {}, "wouldn't": {},
	"you": {}, "you'd": {}, "you'll": {}, "you're": {}, "you've": {}, "your": {}, "yours": {}, "yourself": {}, "yourselves": {},
}

func CreateIndex(useCache bool) {
	if useCache && helpers.FileExists("index.json") {
		idxMu.Lock()
		candidate := newIndex()
		file, err := os.Open("index.json")
		if err == nil {
			decoder := json.NewDecoder(file)
			err = decoder.Decode(candidate)
			file.Close()
			if err == nil && candidate.ready() {
				idx = candidate
				idxMu.Unlock()
				log.Println("Index loaded from cache")
				return
			}
		}
		idxMu.Unlock()

		if err != nil {
			log.Printf("Cache incompatible or unreadable (%v). Rebuilding…", err)
		} else {
			log.Println("Cache incompatible. Rebuilding…")
		}
	}

	start := time.Now()

	idxMu.Lock()
	defer idxMu.Unlock()

	categories, err := persistence.GetCategories()
	if err != nil {
		log.Println("Error getting categories: ", err)
		log.Fatal(err)
	}

	newIdx := newIndex()
	docCount := 0
	for _, category := range categories {
		pages, err := persistence.GetPagesFromCategoryWithContent(category)
		if err != nil {
			log.Println("Error getting pages: ", err)
			log.Fatal(err)
		}
		for _, page := range pages {
			newIdx.addDocument(model.Document{
				Title: page.Title,
				Text:  page.Content,
				ID:    docCount,
				DbId:  page.Id,
			})
			docCount++
		}
	}

	idx = newIdx
	log.Printf("Loaded and indexed %d documents in %v", docCount, time.Since(start))
	start = time.Now()

	file, err := os.Create("index.json")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(idx); err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	log.Printf("Index cache saved in %v", time.Since(start))
}

func AddPage(page model.Page) {
	idxMu.Lock()
	defer idxMu.Unlock()

	idx.addDocument(model.Document{
		Title: page.Title,
		Text:  page.Content,
		DbId:  page.Id,
	})
}

func UpdatePage(page model.Page) {
	idxMu.Lock()
	defer idxMu.Unlock()

	idx.addDocument(model.Document{
		Title: page.Title,
		Text:  page.Content,
		DbId:  page.Id,
	})
}

func DeletePage(id int) {
	idxMu.Lock()
	defer idxMu.Unlock()

	idx.deleteDocument(id)
}

func Search(query string) ([]model.Page, error) {
	const maxItems = 40

	terms := analyze(query)
	if len(terms) == 0 {
		return []model.Page{}, nil
	}

	idxMu.RLock()
	hits := idx.search(terms)
	dbIDs := make([]int, 0, maxItems)
	for i := 0; i < len(hits) && len(dbIDs) < maxItems; i++ {
		docID := hits[i].DocID
		if int(docID) >= len(idx.DocID2DBID) {
			continue
		}
		dbID := idx.DocID2DBID[docID]
		if dbID == 0 {
			continue
		}
		dbIDs = append(dbIDs, dbID)
	}
	idxMu.RUnlock()

	pages := make([]model.Page, 0, len(dbIDs))
	for _, dbID := range dbIDs {
		page, err := persistence.GetPage(dbID)
		if err != nil {
			log.Printf("Page with DB ID %d not found, skipping.", dbID)
			continue
		}
		if page.Id == 0 {
			continue
		}
		pages = append(pages, page)
	}
	return pages, nil
}
