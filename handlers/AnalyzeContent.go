package handlers

import (
	"fmt"
	"net/http"
	"strings"
)

func Analyze(w http.ResponseWriter, r *http.Request) {


}

package main

import (
"fmt"
"strings"

// You will need to run: go get github.com/texttheater/golang-levenshtein/levenshtein
"github.com/texttheater/golang-levenshtein/levenshtein"
)

type Record struct {
	ID      int
	Content string
}

func main() {
	// Sample data with potential overlaps/typos
	record1 := Record{ID: 1, Content: "Data Science with Python"}
	record2 := Record{ID: 2, Content: "Data science with Pythn"}

	// 1. Normalization (Crucial for fuzzy matching)
	// We convert to lowercase and trim whitespace so "Apple" matches "apple"
	str1 := strings.ToLower(strings.TrimSpace(record1.Content))
	str2 := strings.ToLower(strings.TrimSpace(record2.Content))

	// 2. Set a Threshold
	// A threshold of 3 means we allow up to 2 edits (insertions, deletions, or swaps)
	threshold := 3

	// 3. Compute Levenshtein Distance
	// Options can be nil for default weights (1 for ins, del, and sub)
	distance := levenshtein.DistanceForStrings([]rune(str1), []rune(str2), levenshtein.DefaultOptions)

	fmt.Printf("Comparing: '%s' vs '%s'\n", record1.Content, record2.Content)
	fmt.Printf("Edit Distance: %d\n", distance)

	// 4. Determine Overlap
	if distance < threshold {
		fmt.Println("Result: These records likely overlap (Match found).")
	} else {
		fmt.Println("Result: These records are distinct.")
	}
}