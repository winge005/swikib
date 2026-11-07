package search

import (
	"swiki/model"
	"testing"
)

func TestIndexIncrementalAddUpdateDelete(t *testing.T) {
	ix := newIndex()
	ix.addDocument(model.Document{Title: "golang", Text: "golang concurrency", DbId: 1})
	ix.addDocument(model.Document{Title: "javaword", Text: "javaword virtual machine", DbId: 2})

	if got := len(ix.search(analyze("golang"))); got != 1 {
		t.Fatalf("expected 1 golang hit, got %d", got)
	}

	ix.addDocument(model.Document{Title: "rustlang", Text: "rustlang ownership", DbId: 1})

	if got := len(ix.search(analyze("golang"))); got != 0 {
		t.Fatalf("expected old terms removed after update, got %d hits", got)
	}
	if got := len(ix.search(analyze("rustlang"))); got != 1 {
		t.Fatalf("expected 1 rustlang hit after update, got %d", got)
	}
	if ix.N != 2 {
		t.Fatalf("expected 2 live documents after update, got %d", ix.N)
	}

	ix.deleteDocument(2)

	if got := len(ix.search(analyze("javaword"))); got != 0 {
		t.Fatalf("expected deleted document to be removed, got %d hits", got)
	}
	if ix.N != 1 {
		t.Fatalf("expected 1 live document after delete, got %d", ix.N)
	}
}

func TestIndexTitleMatchesRankHigher(t *testing.T) {
	ix := newIndex()
	ix.addDocument(model.Document{Title: "other", Text: "common alpha", DbId: 1})
	ix.addDocument(model.Document{Title: "alpha", Text: "common", DbId: 2})

	hits := ix.search(analyze("alpha"))
	if len(hits) != 2 {
		t.Fatalf("expected 2 hits, got %d", len(hits))
	}
	if hits[0].DocID != 1 {
		t.Fatalf("expected title match to rank first, got docID %d", hits[0].DocID)
	}
}
