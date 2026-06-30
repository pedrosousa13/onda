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
