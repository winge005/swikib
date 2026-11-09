package similarity

import "testing"

func TestJaccardIdentical(t *testing.T) {
	score := Jaccard("Rusland viel Oekraïne binnen", "Rusland viel Oekraïne binnen")
	if score != 1 {
		t.Fatalf("expected 1, got %v", score)
	}
}

func TestJaccardDifferent(t *testing.T) {
	score := Jaccard("Rusland viel Oekraïne binnen", "Het weer is vandaag zonnig")
	if score >= 0.45 {
		t.Fatalf("expected low score, got %v", score)
	}
}

func TestJaccardEmpty(t *testing.T) {
	if got := JaccardSets(map[string]struct{}{}, map[string]struct{}{}); got != 1 {
		t.Fatalf("expected 1 for two empty sets, got %v", got)
	}
}

func TestShingles(t *testing.T) {
	got := Shingles("Rusland viel Oekraïne binnen in februari 2022", 3)
	want := []string{
		"rusland viel oekraïne",
		"viel oekraïne binnen",
		"oekraïne binnen in",
		"binnen in februari",
		"in februari 2022",
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d shingles, got %d", len(want), len(got))
	}

	for _, shingle := range want {
		if _, ok := got[shingle]; !ok {
			t.Fatalf("missing shingle %q", shingle)
		}
	}
}

func TestScorePrefersSimilarContent(t *testing.T) {
	similar := Score(
		"De val van het Romeinse Rijk",
		"Het Romeinse Rijk viel uiteen door politieke instabiliteit en economische problemen.",
		"Waarom Rome uiteindelijk instortte",
		"Het Romeinse Rijk viel uiteen door politieke instabiliteit en economische problemen.",
		false,
	)
	different := Score(
		"De val van het Romeinse Rijk",
		"Het Romeinse Rijk viel uiteen door politieke instabiliteit en economische problemen.",
		"Het weer in Nederland",
		"Vandaag schijnt de zon en blijft het droog.",
		false,
	)

	if similar <= different {
		t.Fatalf("expected similar (%v) > different (%v)", similar, different)
	}
	if similar < 0.45 {
		t.Fatalf("expected similar score above review threshold, got %v", similar)
	}
}

func TestScoreTitleOnly(t *testing.T) {
	score := Score(
		"De val van het Romeinse Rijk",
		"",
		"De val van het Romeinse Rijk",
		"Het Romeinse Rijk viel uiteen door politieke instabiliteit.",
		false,
	)

	if score != 1 {
		t.Fatalf("expected 1 for an identical title, got %v", score)
	}
}

func TestBuildMatchQuery(t *testing.T) {
	query := BuildMatchQuery("Rusland viel Oekraïne binnen in februari 2022", 40)
	if query == "" {
		t.Fatal("expected a query")
	}
	if query != `"februari" OR "oekraïne" OR "rusland" OR "binnen" OR "2022" OR "viel"` {
		t.Fatalf("unexpected query: %s", query)
	}
}

func TestBuildMatchQueryEscapesQuotes(t *testing.T) {
	query := BuildMatchQuery(`zeg "hallo" wereld`, 40)
	if query != `"wereld" OR "hallo" OR "zeg"` {
		t.Fatalf("unexpected query: %s", query)
	}
}
