package similarity

import (
	"regexp"
	"sort"
	"strings"
)

var nonWord = regexp.MustCompile(`[^\p{L}\p{N}]+`)

func normalize(s string) string {
	return nonWord.ReplaceAllString(strings.ToLower(s), " ")
}

func words(s string) []string {
	return strings.Fields(normalize(s))
}

func Tokenize(s string) map[string]struct{} {
	set := make(map[string]struct{})

	for _, word := range words(s) {
		if len([]rune(word)) < 3 {
			continue
		}
		set[word] = struct{}{}
	}

	return set
}

func Shingles(s string, size int) map[string]struct{} {
	set := make(map[string]struct{})

	if size <= 0 {
		return set
	}

	w := words(s)
	if len(w) < size {
		return set
	}

	for i := 0; i <= len(w)-size; i++ {
		set[strings.Join(w[i:i+size], " ")] = struct{}{}
	}

	return set
}

func JaccardSets(a, b map[string]struct{}) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1
	}

	intersection := 0
	for value := range a {
		if _, ok := b[value]; ok {
			intersection++
		}
	}

	union := len(a) + len(b) - intersection
	if union == 0 {
		return 0
	}

	return float64(intersection) / float64(union)
}

func Jaccard(a, b string) float64 {
	return JaccardSets(Tokenize(a), Tokenize(b))
}

func Score(titleA, contentA, titleB, contentB string, sameCategory bool) float64 {
	return ScoreSets(
		Tokenize(titleA),
		Shingles(contentA, 3),
		Tokenize(titleB),
		Shingles(contentB, 3),
		sameCategory,
	)
}

func ScoreSets(titleA, contentA, titleB, contentB map[string]struct{}, sameCategory bool) float64 {
	titleSimilarity := JaccardSets(titleA, titleB)

	contentSimilarity := 0.0
	if len(contentA) > 0 || len(contentB) > 0 {
		contentSimilarity = JaccardSets(contentA, contentB)
	} else {
		contentSimilarity = titleSimilarity
	}

	score := titleSimilarity*0.35 + contentSimilarity*0.65

	if sameCategory {
		score += 0.05
	}

	if score > 1 {
		return 1
	}

	return score
}

func TopTerms(text string, maxTerms int) []string {
	if maxTerms <= 0 {
		maxTerms = 32
	}

	type term struct {
		word  string
		count int
	}

	frequencies := make(map[string]int)
	for _, word := range words(text) {
		if len([]rune(word)) < 3 {
			continue
		}
		frequencies[word]++
	}

	if len(frequencies) == 0 {
		return nil
	}

	terms := make([]term, 0, len(frequencies))
	for word, count := range frequencies {
		terms = append(terms, term{word: word, count: count})
	}

	sort.Slice(terms, func(i, j int) bool {
		if terms[i].count != terms[j].count {
			return terms[i].count > terms[j].count
		}
		if len([]rune(terms[i].word)) != len([]rune(terms[j].word)) {
			return len([]rune(terms[i].word)) > len([]rune(terms[j].word))
		}
		return terms[i].word < terms[j].word
	})

	if len(terms) > maxTerms {
		terms = terms[:maxTerms]
	}

	result := make([]string, len(terms))
	for i, t := range terms {
		result[i] = t.word
	}
	return result
}

func BuildMatchQuery(text string, maxTerms int) string {
	terms := TopTerms(text, maxTerms)
	if len(terms) == 0 {
		return ""
	}

	quoted := make([]string, len(terms))
	for i, term := range terms {
		quoted[i] = `"` + strings.ReplaceAll(term, `"`, `""`) + `"`
	}

	return strings.Join(quoted, " OR ")
}
