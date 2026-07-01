package directory

import (
	"context"
	"testing"
	"time"

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

type fakeOnline struct {
	queried string
	result  []domain.Station
}

func (f *fakeOnline) Search(_ context.Context, q string) ([]domain.Station, error) {
	f.queried = q
	return f.result, nil
}

func TestSearchFallsBackToOnlineWhenNoCorpus(t *testing.T) {
	on := &fakeOnline{result: []domain.Station{{Name: "Radio Eins"}}}
	d := &Directory{Online: on, Offline: NewOffline()} // no corpus loaded
	got, err := d.Search(context.Background(), "radio eins")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if on.queried != "radio eins" {
		t.Fatalf("online source not queried; queried=%q", on.queried)
	}
	if len(got) != 1 || got[0].Name != "Radio Eins" {
		t.Fatalf("want online result, got %+v", got)
	}
}

func TestDirectoryClearCorpus(t *testing.T) {
	d := &Directory{Corpus: NewCorpusStore(t.TempDir(), time.Hour)}
	d.setCorpus([]domain.Station{{Name: "Radio Eins"}})
	if err := d.Corpus.Save(d.snapshot()); err != nil {
		t.Fatalf("save: %v", err)
	}
	if _, ok := d.CorpusSize(); !ok {
		t.Fatal("expected a cached dump before clearing")
	}
	if err := d.ClearCorpus(); err != nil {
		t.Fatalf("ClearCorpus: %v", err)
	}
	if len(d.snapshot()) != 0 {
		t.Fatalf("expected empty in-memory corpus after clearing, got %+v", d.snapshot())
	}
	if _, ok := d.CorpusSize(); ok {
		t.Fatal("expected no cached dump after clearing")
	}
}

func TestSearchUsesCorpusWhenLoaded(t *testing.T) {
	on := &fakeOnline{}
	d := &Directory{Online: on}
	d.setCorpus([]domain.Station{{Name: "Jazz FM"}, {Name: "Rock FM"}})
	got, _ := d.Search(context.Background(), "jazz")
	if on.queried != "" {
		t.Fatal("online source must NOT be queried when corpus is loaded")
	}
	if len(got) == 0 || got[0].Name != "Jazz FM" {
		t.Fatalf("want local corpus match, got %+v", got)
	}
}
