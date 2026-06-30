package directory

import (
	"context"
	"testing"

	"github.com/pedrosousa13/onda/internal/domain"
)

func seededDir() *Directory {
	d := &Directory{}
	d.setCorpus([]domain.Station{
		{Name: "Radio Eins", Country: "Germany", Votes: 5},
		{Name: "Jazz FM", Country: "United Kingdom", Tags: []string{"jazz"}, Votes: 50},
	})
	return d
}

func TestDirectorySearchIsLocalAndFuzzy(t *testing.T) {
	got, err := seededDir().Search(context.Background(), "raido einz")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "Radio Eins" {
		t.Fatalf("local fuzzy search failed: %+v", got)
	}
}

func TestDirectoryPopularSortsByVotes(t *testing.T) {
	got, err := seededDir().Popular(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got[0].Name != "Jazz FM" {
		t.Fatalf("Popular should sort by votes desc, got %+v", got)
	}
}
