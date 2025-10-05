package search

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"swiki/model"
	"swiki/persistence"
	"time"
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

func CreateIndex() {
	var docs []model.Document

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
			fmt.Println(page.Content)
		}
	}

	log.Printf("Loaded %d documents in %v", len(docs), time.Since(start))
	start = time.Now()

	idx := index{}
	idx.add(docs)
	log.Printf("Indexed %d documents in %v", len(docs), time.Since(start))

	start = time.Now()
	matchedIDs := idx.search("caveats")
	log.Printf("Search found %d documents in %v", len(matchedIDs), time.Since(start))

	for _, id := range matchedIDs {
		doc := docs[id]
		log.Printf("%d\t%s\n", id, doc.Title, doc.DbId)
		log.Println("====================================")
	}
}

func Search(query string) ([]model.Page, error) {
	var q string
	var pages []model.Page

	ands, ors, err := tokenizes(query)
	if err != nil {
		if len(query) == 0 {
			return pages, err
		}
		q = " content like('%" + query + "%')"
	}

	for _, str := range ands {
		q += "content like('%" + str + "%') and "
	}

	for _, str := range ors {
		q += "content like('%" + str + "%') or "
	}

	if strings.HasSuffix(q, " and ") {
		q = q[0 : len(q)-5]
	}

	if strings.HasSuffix(q, " or ") {
		q = q[0 : len(q)-4]
	}

	pages, err = persistence.GetPageByQuery(q)
	if err != nil {
		return pages, nil
	}
	return pages, nil
}

func tokenizes(text string) ([]string, []string, error) {
	ands := strings.Split(text, "&&")
	ors := strings.Split(text, "||")
	if len(ands) == 1 && len(ors) == 1 {
		return nil, nil, errors.New("No && or ||")
	}
	if len(ands) > 1 && len(ors) > 1 {
		return nil, nil, errors.New("&& and || used together")
	}

	if len(ands) > 1 {
		ands = trimSlice(ands)
		return ands, nil, nil
	}

	if len(ors) > 1 {
		ors = trimSlice(ors)
		return nil, ors, nil
	}

	return ands, ors, nil
}

func trimSlice(slice []string) []string {
	for i, s := range slice {
		slice[i] = strings.TrimSpace(s)
	}
	return slice
}
