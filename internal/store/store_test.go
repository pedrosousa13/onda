package store

import (
	"path/filepath"
	"testing"

	"github.com/pedrosousa13/onda/internal/domain"
)

func TestConfigDefaults(t *testing.T) {
	c := DefaultConfig()
	if c.Quality != "highest" {
		t.Fatalf("default quality should be highest, got %q", c.Quality)
	}
	if c.Tracking != "never" {
		t.Fatalf("default tracking must be never, got %q", c.Tracking)
	}
	if c.HistoryEnabled {
		t.Fatal("history must default to disabled")
	}
	if c.Normalize {
		t.Fatal("loudness normalization must default to disabled")
	}
	if c.Volume != 100 {
		t.Fatalf("default volume should be 100, got %d", c.Volume)
	}
}

func TestNormalizeRoundTrip(t *testing.T) {
	s := &Store{dir: t.TempDir()}
	if err := s.SaveNormalize(true); err != nil {
		t.Fatal(err)
	}
	got, err := s.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if !got.Normalize {
		t.Fatal("normalize did not round-trip as true")
	}
}

func TestVolumeRoundTrip(t *testing.T) {
	s := &Store{dir: t.TempDir()}
	if err := s.SaveVolume(37); err != nil {
		t.Fatal(err)
	}
	got, err := s.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if got.Volume != 37 {
		t.Fatalf("volume did not round-trip, got %d", got.Volume)
	}
}

func TestConfigRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s := &Store{dir: dir}
	c := DefaultConfig()
	c.Quality = "balanced"
	if err := s.SaveConfig(c); err != nil {
		t.Fatal(err)
	}
	got, err := s.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if got.Quality != "balanced" {
		t.Fatalf("want balanced, got %q", got.Quality)
	}
	if filepath.Base(s.configPath()) != "config.toml" {
		t.Fatal("config should be config.toml")
	}
}

func TestUpdateCheckDefaultAndRoundTrip(t *testing.T) {
	if !DefaultConfig().UpdateCheck {
		t.Fatal("UpdateCheck should default to true (opt-out)")
	}
	s := &Store{dir: t.TempDir()}
	if err := s.SaveUpdateCheck(false); err != nil {
		t.Fatal(err)
	}
	got, err := s.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if got.UpdateCheck {
		t.Fatal("UpdateCheck did not round-trip as false")
	}
}

func TestFavoritesRoundTrip(t *testing.T) {
	s := &Store{dir: t.TempDir()}
	if err := s.AddFavorite(domain.Station{Name: "KEXP", Homepage: "kexp.org"}); err != nil {
		t.Fatal(err)
	}
	_ = s.AddFavorite(domain.Station{Name: "KEXP", Homepage: "kexp.org"})
	favs, err := s.Favorites()
	if err != nil {
		t.Fatal(err)
	}
	if len(favs) != 1 {
		t.Fatalf("want 1 favorite, got %d", len(favs))
	}
	if err := s.RemoveFavorite(domain.Station{Name: "KEXP", Homepage: "kexp.org"}); err != nil {
		t.Fatal(err)
	}
	favs, _ = s.Favorites()
	if len(favs) != 0 {
		t.Fatal("favorite was not removed")
	}
}

func TestCustomStations(t *testing.T) {
	s := &Store{dir: t.TempDir()}
	cs := domain.Station{Name: "My Stream", Variants: []domain.StreamVariant{{URL: "http://x", Bitrate: 128}}}
	if err := s.AddCustom(cs); err != nil {
		t.Fatal(err)
	}
	got, _ := s.CustomStations()
	if len(got) != 1 || got[0].Name != "My Stream" {
		t.Fatalf("custom station not stored: %+v", got)
	}
}
