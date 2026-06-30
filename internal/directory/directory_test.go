package directory

import (
	"context"
	"errors"
	"testing"

	"github.com/pedrosousa13/radio/internal/domain"
)

type fakeSource struct {
	out []domain.Station
	err error
}

func (f fakeSource) Search(context.Context, string) ([]domain.Station, error) {
	return f.out, f.err
}

// fakePopular implements both Search and TopVoted.
type fakePopular struct {
	top []domain.Station
}

func (f fakePopular) Search(context.Context, string) ([]domain.Station, error) {
	return nil, nil
}
func (f fakePopular) TopVoted(context.Context, int) ([]domain.Station, error) {
	return f.top, nil
}

func TestPopularUsesTopVoted(t *testing.T) {
	d := &Directory{
		Online:  fakePopular{top: []domain.Station{{Name: "Top1"}}},
		Offline: fakeSource{out: []domain.Station{{Name: "Bundled"}}},
		Cache:   NewCache(t.TempDir(), 0),
	}
	got, err := d.Popular(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "Top1" {
		t.Fatalf("expected top-voted result, got %+v", got)
	}
}

func TestPopularFallsBackToOfflineWithoutTopVoted(t *testing.T) {
	// fakeSource has no TopVoted method → Popular must fall back to offline.
	d := &Directory{
		Online:  fakeSource{out: nil},
		Offline: fakeSource{out: []domain.Station{{Name: "Bundled"}}},
		Cache:   NewCache(t.TempDir(), 0),
	}
	got, err := d.Popular(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "Bundled" {
		t.Fatalf("expected offline fallback, got %+v", got)
	}
}

func TestDirectoryFallsBackToOfflineOnError(t *testing.T) {
	d := &Directory{
		Online:  fakeSource{err: errors.New("network down")},
		Offline: fakeSource{out: []domain.Station{{Name: "Bundled"}}},
		Cache:   NewCache(t.TempDir(), 0),
	}
	got, err := d.Search(context.Background(), "x")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "Bundled" {
		t.Fatalf("expected offline fallback, got %+v", got)
	}
}

func TestDirectoryPrefersOnlineAndCaches(t *testing.T) {
	cache := NewCache(t.TempDir(), 1<<62) // effectively never expires
	d := &Directory{
		Online:  fakeSource{out: []domain.Station{{Name: "Live"}}},
		Offline: fakeSource{out: []domain.Station{{Name: "Bundled"}}},
		Cache:   cache,
	}
	got, _ := d.Search(context.Background(), "x")
	if got[0].Name != "Live" {
		t.Fatalf("expected online result, got %+v", got)
	}
	if cached, ok := cache.Get("x"); !ok || cached[0].Name != "Live" {
		t.Fatal("online result should have been cached")
	}
}
