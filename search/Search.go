package search

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
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

	if useCache {
		if helpers.FileExists("index.json") {
			file, err := os.Open("index.json")
			if err != nil {
				fmt.Println("Error opening file:", err)
				return
			}
			defer file.Close()

			decoder := json.NewDecoder(file)
			if err := decoder.Decode(&idx); err != nil {
				fmt.Println("Error decoding JSON:", err)
				return
			}

			log.Println("Index loaded from cache")
			return
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
			doc := model.Document{Title: page.Title, Text: page.Content, ID: len(docs), DbId: page.Id}
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
	encoder.SetIndent("", "  ") // pretty print (optional)
	if err := encoder.Encode(idx); err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	fmt.Println("✅ JSON index saved")
}

func Search(query string) ([]model.Page, error) {

	var pages []model.Page
	matchedIDs := idx.search(query)
	maxItems := 40
	currentItem := 0
	log.Printf("Search found %d documents", len(matchedIDs))

	for _, id := range matchedIDs {
		currentItem = currentItem + 1
		page, err := persistence.GetPage(id)
		if err != nil {
			log.Printf("Page with ID %d not found, skipping.", id)
			continue
		}
		pages = append(pages, page)
		if currentItem >= maxItems {
			break
		}
	}
	return pages, nil
}
