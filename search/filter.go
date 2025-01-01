package search

import (
	snowballeng "github.com/kljensen/snowball/english"
	"strings"
)

func lowercaseFilter(tokens []string) []string {
	r := make([]string, len(tokens))
	for i, token := range tokens {
		r[i] = strings.ToLower(token)
	}
	return r
}

func stopwordFilter(tokens []string) []string {
	r := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if _, ok := stopwords[token]; !ok {
			r = append(r, token)
		}
	}
	return r
}

func stemmerFilter(tokens []string) []string {
	r := make([]string, len(tokens))
	for i, token := range tokens {
		r[i] = snowballeng.Stem(token, false)
	}
	return r
}
