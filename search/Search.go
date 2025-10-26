package search

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"swiki/helpers"
	"swiki/model"
	"swiki/persistence"
	"time"
)

var docs []model.Document
var idx = index{}

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
	docs = nil
	idx = index{}

	if useCache && helpers.FileExists("index.json") {
		file, err := os.Open("index.json")
		if err == nil {
			defer file.Close()
			decoder := json.NewDecoder(file)
			if err := decoder.Decode(&idx); err == nil && idx.Postings != nil && idx.N > 0 {
				log.Println("Index loaded from cache")
				return
			}
			// If we get here, the cache is incompatible or broken.
			log.Printf("Cache incompatible or unreadable (%v). Rebuilding…", err)
		} else {
			log.Printf("Could not open cache: %v. Rebuilding…", err)
		}
	}

	start := time.Now()

	categories, err := persistence.GetCategories()
	if err != nil {
		log.Println("Error getting categories: ", err)
		log.Fatal(err)
	}

	for _, category := range categories {
		pages, err := persistence.GetPagesFromCategoryWithContent(category)
		if err != nil {
			log.Println("Error getting categories: ", err)
			log.Fatal(err)
		}
		for _, page := range pages {
			doc := model.Document{
				Title: page.Title,
				Text:  page.Content,
				ID:    len(docs),
				DbId:  page.Id,
			}
			docs = append(docs, doc)
		}
	}

	log.Printf("Loaded %d documents in %v", len(docs), time.Since(start))
	start = time.Now()

	idx.add(docs)
	log.Printf("Indexed %d documents in %v", len(docs), time.Since(start))

	file, err := os.Create("index.json")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(idx); err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	fmt.Println("✅ JSON index saved")
}

func Search(query string) ([]model.Page, error) {
	const maxItems = 40

	// Optional: if user typed a quoted phrase, we boost those docs later
	var phraseBoost map[uint32]bool
	if strings.Contains(query, "\"") {
		if terms := parseQuotedTerms(query); len(terms) > 0 {
			ids := idx.phraseSearch(terms)
			if len(ids) > 0 {
				phraseBoost = make(map[uint32]bool, len(ids))
				for _, id := range ids {
					phraseBoost[id] = true
				}
			}
		}
	}

	// Ranked search with BM25
	hits := idx.bm25Search(query)
	if len(hits) == 0 {
		return nil, nil
	}

	// Boost phrase matches (if any), then sort again
	if phraseBoost != nil {
		for i := range hits {
			if phraseBoost[hits[i].DocID] {
				hits[i].Score *= 1.5
			}
		}
		sort.Slice(hits, func(i, j int) bool { return hits[i].Score > hits[j].Score })
	}

	// Map internal DocID -> real DB ID using the new mapping (fixes DocID!=DBID)
	pages := make([]model.Page, 0, maxItems)
	for i := 0; i < len(hits) && len(pages) < maxItems; i++ {
		docID := hits[i].DocID
		if int(docID) >= len(idx.DocID2DBID) {
			continue
		}
		dbID := idx.DocID2DBID[docID]
		page, err := persistence.GetPage(dbID)
		if err != nil {
			log.Printf("Page with DB ID %d not found, skipping.", dbID)
			continue
		}
		pages = append(pages, page)
	}
	return pages, nil
}
