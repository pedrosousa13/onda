package store

import (
	"path/filepath"
	"testing"
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
