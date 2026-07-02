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

// fakeFetcher is an online Source that can also serve a full dump, so it
// satisfies the fullFetcherProgress path used by RefreshWithProgress.
type fakeFetcher struct{ stations []domain.Station }

func (f *fakeFetcher) Search(context.Context, string) ([]domain.Station, error) { return nil, nil }
func (f *fakeFetcher) FetchAllWithProgress(context.Context, func(int64)) ([]domain.Station, error) {
	return f.stations, nil
}

func nStations(n int) []domain.Station {
	out := make([]domain.Station, n)
	for i := range out {
		out[i] = domain.Station{Name: "s"}
	}
	return out
}

func TestRefreshKeepsCorpusWhenDumpTooSmall(t *testing.T) {
	d := &Directory{
		Online: &fakeFetcher{stations: nStations(minPlausibleCorpus - 1)},
		Corpus: NewCorpusStore(t.TempDir(), time.Hour),
	}
	d.setCorpus(nStations(3000)) // a good, existing corpus
	if _, err := d.RefreshWithProgress(context.Background(), nil); err == nil {
		t.Fatal("expected an incomplete dump to be rejected")
	}
	if len(d.snapshot()) != 3000 {
		t.Fatalf("existing corpus must survive a rejected refresh, got %d", len(d.snapshot()))
	}
	if _, ok := d.CorpusSize(); ok {
		t.Fatal("a rejected dump must not be persisted")
	}
}

func TestRefreshReplacesCorpusWithFullDump(t *testing.T) {
	d := &Directory{
		Online: &fakeFetcher{stations: nStations(minPlausibleCorpus)},
		Corpus: NewCorpusStore(t.TempDir(), time.Hour),
	}
	out, err := d.RefreshWithProgress(context.Background(), nil)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if len(out) != minPlausibleCorpus || len(d.snapshot()) != minPlausibleCorpus {
		t.Fatalf("corpus should be replaced by the full dump, got %d", len(d.snapshot()))
	}
	if _, ok := d.CorpusSize(); !ok {
		t.Fatal("a full dump should be persisted")
	}
}
