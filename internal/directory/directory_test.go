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
