package directory

import (
	"context"
	"testing"

	"github.com/pedrosousa13/onda/internal/domain"
)

// fakeSource is a Source test double that returns a fixed station list.
type fakeSource struct {
	out []domain.Station
	err error
}

func (f fakeSource) Search(_ context.Context, _ string) ([]domain.Station, error) {
	return f.out, f.err
}

func browseFixture() []domain.Station {
	return []domain.Station{
		{Name: "Radio Eins", Country: "Germany", Language: "German", Tags: []string{"pop", "talk"}, Votes: 5, Trend: 0},
		{Name: "Jazz FM", Country: "United Kingdom", Language: "English", Tags: []string{"jazz"}, Votes: 50, Trend: 3},
		{Name: "Rock FM", Country: "United Kingdom", Language: "English", Tags: []string{"rock", "talk"}, Votes: 20, Trend: 1},
		{Name: "Antenne", Country: "Germany", Language: "German", Tags: []string{"pop"}, Votes: 30, Trend: 0},
	}
}

func browseFixtureDir() *Directory {
	return &Directory{Offline: fakeSource{out: browseFixture()}}
}

func TestDirectoryCountries(t *testing.T) {
	facets, err := browseFixtureDir().Countries(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := []domain.Facet{
		{Name: "Germany", Count: 2},
		{Name: "United Kingdom", Count: 2},
	}
	if len(facets) != len(want) {
		t.Fatalf("got %+v, want %+v", facets, want)
	}
	for i, f := range facets {
		if f != want[i] {
			t.Fatalf("got %+v, want %+v", facets, want)
		}
	}
}

func TestDirectoryLanguages(t *testing.T) {
	facets, err := browseFixtureDir().Languages(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	want := []domain.Facet{
		{Name: "English", Count: 2},
		{Name: "German", Count: 2},
	}
	if len(facets) != len(want) {
		t.Fatalf("got %+v, want %+v", facets, want)
	}
	for i, f := range facets {
		if f != want[i] {
			t.Fatalf("got %+v, want %+v", facets, want)
		}
	}
}

func TestDirectoryTags(t *testing.T) {
	facets, err := browseFixtureDir().Tags(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	// pop: Radio Eins, Antenne = 2; talk: Radio Eins, Rock FM = 2; jazz: 1; rock: 1
	// count-desc, tie name-asc: pop(2) < talk(2) alphabetically -> pop, talk, jazz, rock
	want := []domain.Facet{
		{Name: "pop", Count: 2},
		{Name: "talk", Count: 2},
		{Name: "jazz", Count: 1},
		{Name: "rock", Count: 1},
	}
	if len(facets) != len(want) {
		t.Fatalf("got %+v, want %+v", facets, want)
	}
	for i, f := range facets {
		if f != want[i] {
			t.Fatalf("got %+v, want %+v", facets, want)
		}
	}
}

func TestDirectoryTagsCapsAt100(t *testing.T) {
	stations := make([]domain.Station, 0, 150)
	for i := 0; i < 150; i++ {
		stations = append(stations, domain.Station{
			Name: "Station", Tags: []string{"tag" + string(rune('A'+i%26)) + string(rune('a'+i/26))},
		})
	}
	d := &Directory{Offline: fakeSource{out: stations}}
	facets, err := d.Tags(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(facets) > 100 {
		t.Fatalf("tagFacets must cap at 100, got %d", len(facets))
	}
}

func TestDirectoryStationsByCountryCaseInsensitive(t *testing.T) {
	got, err := browseFixtureDir().StationsBy(context.Background(), domain.AxisCountry, "germany", domain.Sort{Key: domain.SortVotes})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 German stations, got %+v", got)
	}
	// votes desc default: Antenne(30) before Radio Eins(5)
	if got[0].Name != "Antenne" || got[1].Name != "Radio Eins" {
		t.Fatalf("want votes-desc order, got %+v", got)
	}
}

func TestDirectoryStationsBySortNameAsc(t *testing.T) {
	got, err := browseFixtureDir().StationsBy(context.Background(), domain.AxisCountry, "United Kingdom", domain.Sort{Key: domain.SortName})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 UK stations, got %+v", got)
	}
	if got[0].Name != "Jazz FM" || got[1].Name != "Rock FM" {
		t.Fatalf("want name-asc order, got %+v", got)
	}
}

func TestDirectoryStationsBySortVotesFlipped(t *testing.T) {
	got, err := browseFixtureDir().StationsBy(context.Background(), domain.AxisCountry, "United Kingdom", domain.Sort{Key: domain.SortVotes, Flip: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 UK stations, got %+v", got)
	}
	// default votes is desc; Flip=true reverses to asc: Rock FM(20) before Jazz FM(50)
	if got[0].Name != "Rock FM" || got[1].Name != "Jazz FM" {
		t.Fatalf("want votes-asc (flipped) order, got %+v", got)
	}
}

func TestDirectoryStationsByTag(t *testing.T) {
	got, err := browseFixtureDir().StationsBy(context.Background(), domain.AxisTag, "TALK", domain.Sort{Key: domain.SortVotes})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 stations tagged talk, got %+v", got)
	}
	if got[0].Name != "Rock FM" || got[1].Name != "Radio Eins" {
		t.Fatalf("want votes-desc order, got %+v", got)
	}
}

func TestDirectoryStationsByLanguage(t *testing.T) {
	got, err := browseFixtureDir().StationsBy(context.Background(), domain.AxisLanguage, "english", domain.Sort{Key: domain.SortVotes})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("want 2 English stations, got %+v", got)
	}
	if got[0].Name != "Jazz FM" || got[1].Name != "Rock FM" {
		t.Fatalf("want votes-desc order, got %+v", got)
	}
}
