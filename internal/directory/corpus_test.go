package directory

import (
	"os"
	"testing"
	"time"

	"github.com/pedrosousa13/onda/internal/domain"
)

func TestCorpusStoreRoundTrip(t *testing.T) {
	s := NewCorpusStore(t.TempDir(), 7*24*time.Hour)
	in := []domain.Station{{Name: "Radio Eins", Country: "Germany", Votes: 42}}
	if err := s.Save(in); err != nil {
		t.Fatalf("save: %v", err)
	}
	out, at, ok := s.Load()
	if !ok || len(out) != 1 || out[0].Name != "Radio Eins" {
		t.Fatalf("load round-trip failed: ok=%v out=%+v", ok, out)
	}
	if !s.Fresh(at) {
		t.Fatal("just-saved corpus should be fresh")
	}
	if s.Fresh(at.Add(-8 * 24 * time.Hour)) {
		t.Fatal("corpus older than TTL should be stale")
	}
}

func TestCorpusStoreDeleteAndSize(t *testing.T) {
	s := NewCorpusStore(t.TempDir(), time.Hour)
	in := []domain.Station{{Name: "Radio Eins", Country: "Germany", Votes: 42}}
	if err := s.Save(in); err != nil {
		t.Fatalf("save: %v", err)
	}
	if n, ok := s.Size(); !ok || n <= 0 {
		t.Fatalf("expected a positive size for saved corpus, got n=%d ok=%v", n, ok)
	}
	if err := s.Delete(); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, _, ok := s.Load(); ok {
		t.Fatal("deleted corpus should not load")
	}
	if n, ok := s.Size(); ok || n != 0 {
		t.Fatalf("deleted corpus should report no size, got n=%d ok=%v", n, ok)
	}
}

func TestCorpusStoreDeleteMissingIsNotAnError(t *testing.T) {
	s := NewCorpusStore(t.TempDir(), time.Hour)
	if err := s.Delete(); err != nil {
		t.Fatalf("deleting a missing corpus should not error, got %v", err)
	}
}

func TestCorpusStoreMissingAndCorrupt(t *testing.T) {
	dir := t.TempDir()
	s := NewCorpusStore(dir, time.Hour)
	if _, _, ok := s.Load(); ok {
		t.Fatal("missing corpus should not load")
	}
	if err := os.WriteFile(s.path, []byte("not gzip"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, ok := s.Load(); ok {
		t.Fatal("corrupt corpus should not load")
	}
}
