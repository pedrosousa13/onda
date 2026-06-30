package directory

import (
	"testing"
	"time"

	"github.com/pedrosousa13/radio/internal/domain"
)

func TestCacheRoundTrip(t *testing.T) {
	dir := t.TempDir()
	c := NewCache(dir, time.Hour)
	want := []domain.Station{{Name: "KEXP"}}

	if _, ok := c.Get("q"); ok {
		t.Fatal("expected miss on empty cache")
	}
	if err := c.Put("q", want); err != nil {
		t.Fatal(err)
	}
	got, ok := c.Get("q")
	if !ok || len(got) != 1 || got[0].Name != "KEXP" {
		t.Fatalf("round-trip failed: %+v ok=%v", got, ok)
	}
}

func TestCacheExpiryWithStaleFallback(t *testing.T) {
	dir := t.TempDir()
	c := NewCache(dir, -time.Second) // everything immediately stale
	_ = c.Put("q", []domain.Station{{Name: "Old"}})

	if _, ok := c.Get("q"); ok {
		t.Fatal("expired entry must be a miss for Get")
	}
	stale, ok := c.Stale("q")
	if !ok || stale[0].Name != "Old" {
		t.Fatal("Stale must still return expired data")
	}
}
