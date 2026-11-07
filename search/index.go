package search

import (
	"math"
	"slices"
	"sort"
	"strings"
	"swiki/model"
	"unicode"
)

type Posting struct {
	DocID     uint32 `json:"d"`
	Freq      uint32 `json:"f"`
	TitleFreq uint32 `json:"tf,omitempty"`
}

type index struct {
	Postings   map[string][]Posting `json:"postings"`
	N          uint32               `json:"N"`
	DocID2DBID []int                `json:"docid2dbid"`
	DBID2DocID map[int]uint32       `json:"dbid2docid"`
	DocTerms   [][]string           `json:"doc_terms"`
}

type hit struct {
	DocID uint32
	Score float64
}

type termCount struct {
	freq      uint32
	titleFreq uint32
}

func newIndex() *index {
	return &index{
		Postings:   make(map[string][]Posting, 1<<15),
		DBID2DocID: make(map[int]uint32),
	}
}

func (ix *index) ready() bool {
	return ix != nil &&
		ix.Postings != nil &&
		ix.DBID2DocID != nil &&
		len(ix.DocID2DBID) == len(ix.DocTerms)
}

func (ix *index) add(docs []model.Document) {
	for _, d := range docs {
		ix.addDocument(d)
	}
}

func (ix *index) addDocument(d model.Document) {
	if docID, ok := ix.DBID2DocID[d.DbId]; ok {
		ix.replaceDocument(docID, d)
		return
	}

	docID := uint32(len(ix.DocID2DBID))
	ix.DocID2DBID = append(ix.DocID2DBID, d.DbId)
	ix.DocTerms = append(ix.DocTerms, nil)
	ix.DBID2DocID[d.DbId] = docID
	ix.indexDocument(docID, d)
}

func (ix *index) replaceDocument(docID uint32, d model.Document) {
	ix.removeDocument(docID)
	ix.DocID2DBID[docID] = d.DbId
	ix.DBID2DocID[d.DbId] = docID
	ix.indexDocument(docID, d)
}

func (ix *index) deleteDocument(dbID int) {
	docID, ok := ix.DBID2DocID[dbID]
	if !ok {
		return
	}
	ix.removeDocument(docID)
}

func (ix *index) indexDocument(docID uint32, d model.Document) {
	titleTerms := analyze(d.Title)
	textTerms := analyze(d.Text)

	counts := make(map[string]termCount, len(titleTerms)+len(textTerms))
	for _, term := range titleTerms {
		c := counts[term]
		c.freq++
		c.titleFreq++
		counts[term] = c
	}
	for _, term := range textTerms {
		c := counts[term]
		c.freq++
		counts[term] = c
	}

	terms := make([]string, 0, len(counts))
	for term, c := range counts {
		ix.Postings[term] = append(ix.Postings[term], Posting{
			DocID:     docID,
			Freq:      c.freq,
			TitleFreq: c.titleFreq,
		})
		terms = append(terms, term)
	}
	sort.Strings(terms)

	ix.DocTerms[docID] = terms
	ix.N++
}

func (ix *index) removeDocument(docID uint32) {
	if int(docID) >= len(ix.DocID2DBID) {
		return
	}

	for _, term := range ix.DocTerms[docID] {
		postings := ix.Postings[term]
		postings = slices.DeleteFunc(postings, func(p Posting) bool {
			return p.DocID == docID
		})
		if len(postings) == 0 {
			delete(ix.Postings, term)
		} else {
			ix.Postings[term] = postings
		}
	}

	if dbID := ix.DocID2DBID[docID]; dbID != 0 {
		delete(ix.DBID2DocID, dbID)
	}

	ix.DocID2DBID[docID] = 0
	ix.DocTerms[docID] = nil
	if ix.N > 0 {
		ix.N--
	}
}

func (ix *index) search(terms []string) []hit {
	seen := make(map[string]struct{}, len(terms))
	scores := make(map[uint32]float64, 256)

	for _, term := range terms {
		if _, ok := seen[term]; ok {
			continue
		}
		seen[term] = struct{}{}

		for _, p := range ix.Postings[term] {
			score := 1.0
			score += 0.5 * math.Log1p(float64(p.Freq))
			score += 1.5 * math.Log1p(float64(p.TitleFreq))
			scores[p.DocID] += score
		}
	}

	out := make([]hit, 0, len(scores))
	for docID, score := range scores {
		out = append(out, hit{DocID: docID, Score: score})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].DocID < out[j].DocID
	})
	return out
}

func tokenize(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
}

func analyze(text string) []string {
	tokens := tokenize(text)
	tokens = lowercaseFilter(tokens)
	tokens = stopwordFilter(tokens)
	tokens = stemmerFilter(tokens)
	return tokens
}
