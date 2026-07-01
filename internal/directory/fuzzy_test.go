package directory

import (
	"testing"

	"github.com/pedrosousa13/onda/internal/domain"
)

func TestRankByQuery(t *testing.T) {
	stations := []domain.Station{
		{Name: "Some Other Station", Country: "France", Tags: []string{"pop"}},
		{Name: "FIP Jazz", Country: "France", Tags: []string{"jazz"}},
		{Name: "Jazz24", Country: "United States", Tags: []string{"jazz"}},
	}
	ranked := rankByQuery("jazz", stations)
	// A name containing "jazz" should outrank a tag-only match, which outranks none.
	if ranked[0].Name != "Jazz24" && ranked[0].Name != "FIP Jazz" {
		t.Fatalf("expected a jazz-named station first, got %q", ranked[0].Name)
	}
	if ranked[len(ranked)-1].Name != "Some Other Station" {
		t.Fatalf("non-matching station should rank last, got %q", ranked[len(ranked)-1].Name)
	}
}

func TestRankByQueryEmpty(t *testing.T) {
	in := []domain.Station{{Name: "B"}, {Name: "A"}}
	out := rankByQuery("", in)
	if len(out) != 2 || out[0].Name != "B" {
		t.Fatal("empty query should preserve order")
	}
}

func TestFieldScoreOrder(t *testing.T) {
	if fieldScore("jazz", "jazz") <= fieldScore("jazz", "smooth jazz") {
		t.Fatal("exact match should outscore substring")
	}
	if fieldScore("jazz", "smooth jazz") <= fieldScore("jazz", "jzz pattern xa") {
		t.Fatal("substring should outscore subsequence")
	}
}

func TestFieldScoreTypoTolerance(t *testing.T) {
	// A transposition typo within edit distance 2 matches a token via the
	// edit-distance tier (raido → radio is distance 1).
	if fieldScore("raido", "Radio Paradise") == 0 {
		t.Fatal("expected 'raido' to fuzzy-match token 'radio'")
	}
	// Transpositions in tag-like single words match too.
	if fieldScore("rcok", "rock") == 0 {
		t.Fatal("expected 'rcok' to fuzzy-match 'rock'")
	}
	// Multi-word typo queries match per token: every query word must fuzzy-match
	// some field word ("raido einz" → "Radio Eins").
	if fieldScore("raido einz", "Radio Eins") == 0 {
		t.Fatal("expected 'raido einz' to fuzzy-match 'Radio Eins'")
	}
	// A short word in a multi-word query must match exactly (no loose matching).
	if fieldScore("raido fm", "radio fm") == 0 {
		t.Fatal("expected 'raido fm' to match 'radio fm' (typo word + exact short word)")
	}
	// If any query word matches nothing, the whole query doesn't fuzzy-match.
	if got := fieldScore("raido zzzzz", "radio eins"); got != 0 {
		t.Fatalf("query with an unmatched word should not fuzzy-match, got %d", got)
	}
	// The edit-distance tier must rank below the substring tier.
	if fieldScore("raido", "Radio Paradise") >= fieldScore("radio", "radio paradise") {
		t.Fatal("edit-distance match should score below a substring match")
	}
}

func TestFieldScoreShortQueryGated(t *testing.T) {
	// "oct" is within edit distance 1 of "cot" (transposition) but is only 3
	// runes, so the length-gated edit-distance tier must not fire — and it is
	// neither a substring nor a subsequence, so the score is 0.
	if got := fieldScore("oct", "cot"); got != 0 {
		t.Fatalf("3-rune query must not use the edit-distance tier, got %d", got)
	}
}

func TestFieldScoreFarTypoNoMatch(t *testing.T) {
	// Distance greater than 2 must not match.
	if got := fieldScore("raidooo", "radio"); got != 0 {
		t.Fatalf("distance>2 should not match, got %d", got)
	}
}

func TestMatchLocalFiltersAndFuzzyMatches(t *testing.T) {
	corpus := []domain.Station{
		{Name: "Radio Eins", Country: "Germany", Tags: []string{"pop"}},
		{Name: "Jazz FM", Country: "United Kingdom", Tags: []string{"jazz"}},
		{Name: "Totally Unrelated", Country: "Chile"},
	}
	// Typo query must surface the right station and drop non-matches.
	got := matchLocal("raido einz", corpus)
	if len(got) != 1 || got[0].Name != "Radio Eins" {
		t.Fatalf("expected only Radio Eins, got %+v", got)
	}
	// Tag match works.
	if got := matchLocal("jazz", corpus); len(got) != 1 || got[0].Name != "Jazz FM" {
		t.Fatalf("expected Jazz FM by tag, got %+v", got)
	}
	// Empty query returns everything unchanged.
	if got := matchLocal("", corpus); len(got) != len(corpus) {
		t.Fatalf("empty query should return all, got %d", len(got))
	}
}
