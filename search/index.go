package search

import (
	"math"
	"sort"
	"strings"
	"swiki/model"
	"unicode"
)

// ---- Data structures ----

type Posting struct {
	DocID     uint32   `json:"d"`
	Freq      uint32   `json:"f"`
	Positions []uint32 `json:"p,omitempty"` // needed for phrase/proximity search (optional)
}

// index: inverted index + corpus stats + internal->DB id mapping
type index struct {
	Postings   map[string][]Posting `json:"postings"`   // term -> postings
	N          uint32               `json:"N"`          // number of docs
	DocLen     []uint32             `json:"doc_len"`    // docID -> token count
	DF         map[string]uint32    `json:"df"`         // term -> document frequency
	AvgDL      float64              `json:"avg_dl"`     // average doc length
	DocID2DBID []int                `json:"docid2dbid"` // internal docID -> DB page ID
}

// NOTE: we index Title + Text so titles get some chance to match strongly.
func (ix *index) add(docs []model.Document) {
	if ix.Postings == nil {
		ix.Postings = make(map[string][]Posting, 1<<15)
	}
	if ix.DF == nil {
		ix.DF = make(map[string]uint32, 1<<15)
	}

	ix.DocLen = make([]uint32, len(docs))
	ix.DocID2DBID = make([]int, len(docs))

	for i, d := range docs {
		terms := analyze(d.Title + " " + d.Text)
		ix.DocLen[i] = uint32(len(terms))
		ix.DocID2DBID[i] = d.DbId // from your Document build in Search.go

		seen := make(map[string]bool, 64)

		for pos, term := range terms {
			pl := ix.Postings[term]
			// append new posting if this doc isn't yet the tail of the list
			if len(pl) == 0 || pl[len(pl)-1].DocID != uint32(i) {
				pl = append(pl, Posting{DocID: uint32(i)})
			}
			// update tail
			p := &pl[len(pl)-1]
			p.Freq++
			p.Positions = append(p.Positions, uint32(pos))
			pl[len(pl)-1] = *p
			ix.Postings[term] = pl

			if !seen[term] {
				ix.DF[term]++
				seen[term] = true
			}
		}
	}
	ix.N = uint32(len(docs))
	var sum uint64
	for _, L := range ix.DocLen {
		sum += uint64(L)
	}
	if ix.N > 0 {
		ix.AvgDL = float64(sum) / float64(ix.N)
	}
}

// ---- Ranked search (BM25) ----

type hit struct {
	DocID uint32
	Score float64
}

// bm25Search runs a standard BM25 ranking over the query terms.
func (ix *index) bm25Search(query string) []hit {
	qterms := analyze(query)
	if len(qterms) == 0 {
		return nil
	}

	// de-duplicate query terms (simple)
	seen := make(map[string]struct{}, len(qterms))
	tmp := qterms[:0]
	for _, t := range qterms {
		if _, ok := seen[t]; !ok {
			seen[t] = struct{}{}
			tmp = append(tmp, t)
		}
	}
	qterms = tmp

	const k1 = 1.2
	const b = 0.75

	scores := make(map[uint32]float64, 256)

	for _, term := range qterms {
		df := float64(ix.DF[term])
		if df == 0 {
			continue
		}
		idf := math.Log((float64(ix.N)-df+0.5)/(df+0.5) + 1.0)

		for _, p := range ix.Postings[term] {
			tf := float64(p.Freq)
			dl := float64(ix.DocLen[p.DocID])
			den := tf + k1*(1.0-b+b*dl/ix.AvgDL)
			score := idf * (tf * (k1 + 1.0) / den)

			// light early-position boost (often captures title/heading/first tokens)
			if len(p.Positions) > 0 && p.Positions[0] < 10 {
				score *= 1.15
			}
			scores[p.DocID] += score
		}
	}

	out := make([]hit, 0, len(scores))
	for id, s := range scores {
		out = append(out, hit{DocID: id, Score: s})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Score > out[j].Score })
	return out
}

// ---- (Optional) phrase support helpers ----
// Keep these only if you want quoted-phrase boosting later.

func tokenize(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		// Split on any character that is not a letter or a number.
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
}

func parseQuotedTerms(q string) []string {
	i := strings.IndexByte(q, '"')
	if i < 0 {
		return nil
	}
	j := strings.IndexByte(q[i+1:], '"')
	if j < 0 {
		return nil
	}
	j = i + 1 + j
	return analyze(q[i+1 : j])
}

func intersectPhrase(a, b []Posting, gap uint32) []Posting {
	i, j := 0, 0
	out := make([]Posting, 0)
	for i < len(a) && j < len(b) {
		if a[i].DocID == b[j].DocID {
			pa, pb := a[i].Positions, b[j].Positions
			x, y := 0, 0
			found := false
			for x < len(pa) && y < len(pb) {
				target := pa[x] + gap
				if pb[y] == target {
					found = true
					break
				}
				if pb[y] < target {
					y++
				} else {
					x++
				}
			}
			if found {
				out = append(out, Posting{DocID: a[i].DocID})
			}
			i++
			j++
		} else if a[i].DocID < b[j].DocID {
			i++
		} else {
			j++
		}
	}
	return out
}

func (ix *index) phraseSearch(terms []string) []uint32 {
	if len(terms) == 0 {
		return nil
	}
	lists := make([][]Posting, 0, len(terms))
	for _, t := range terms {
		pl := ix.Postings[t]
		if len(pl) == 0 {
			return nil
		}
		lists = append(lists, pl)
	}
	cur := lists[0]
	for k := 1; k < len(lists); k++ {
		cur = intersectPhrase(cur, lists[k], 1) // exact adjacency
		if len(cur) == 0 {
			return nil
		}
	}
	out := make([]uint32, len(cur))
	for i, p := range cur {
		out[i] = p.DocID
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// ---- Analyzer pipeline (kept as in your code) ----
// We reuse your lowercaseFilter, stopwordFilter, and stemmerFilter from filter.go.

func analyze(text string) []string {
	tokens := tokenize(text)
	tokens = lowercaseFilter(tokens)
	tokens = stopwordFilter(tokens)
	tokens = stemmerFilter(tokens)
	return tokens
}
