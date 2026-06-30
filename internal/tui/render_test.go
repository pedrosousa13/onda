package tui

import (
	"fmt"
	"strings"
	"testing"

	"github.com/pedrosousa13/onda/internal/domain"
	"github.com/pedrosousa13/onda/internal/update"
)

// sampleModel builds a populated model for visual inspection of View() output.
func sampleModel() Model {
	m := New(nil, nil, nil, domain.QualityHighest, "never", false, "catppuccin-mocha", true, "1.0.0", "")
	m.width, m.height = 76, 22
	m.stations = []domain.Station{
		{Name: "KEXP", Country: "United States", Tags: []string{"indie", "seattle"}, Variants: []domain.StreamVariant{{Bitrate: 128}}},
		{Name: "BBC World Service", Country: "United Kingdom", Tags: []string{"news"}, Variants: []domain.StreamVariant{{Bitrate: 96}}},
		{Name: "FIP", Country: "France", Tags: []string{"eclectic", "jazz"}, Variants: []domain.StreamVariant{{URL: "h", Lossless: true}, {URL: "a", Bitrate: 192}, {URL: "b", Bitrate: 128}}},
		{Name: "NTS Radio 1", Country: "United Kingdom", Tags: []string{"electronic"}, Variants: []domain.StreamVariant{{Bitrate: 128}}},
		{Name: "Radio Nacional de España", Country: "Spain", Variants: []domain.StreamVariant{{Bitrate: 64}}},
		{Name: "Triple J", Country: "Australia", Tags: []string{"alternative"}, Variants: []domain.StreamVariant{{Bitrate: 128}}},
	}
	m.cursor = 2
	m.loading = false
	m.crumb = "popular"
	m.favKeys = map[string]bool{favKey(m.stations[0]): true}
	m.playing = m.stations[2]
	m.isPlaying = true
	m.nowTitle = "Khruangbin — Maria También"
	m.volume = 80
	return m
}

// TestRenderGallery prints each screen so the layout can be reviewed.
// Run: go test ./internal/tui/ -run TestRenderGallery -v
func TestRenderGallery(t *testing.T) {
	frame := func(label string, m Model) {
		fmt.Printf("\n┌──────── %s ", label)
		for i := len(label); i < 60; i++ {
			fmt.Print("─")
		}
		fmt.Println("┐")
		fmt.Println(m.View())
		fmt.Println("└" + repeat("─", 70) + "┘")
	}

	m := sampleModel()
	frame("HOME", m) // sampleModel has a favorite, so Home shows favorites

	bw := m
	bw.view = viewBrowse
	bw.crumb = "popular"
	frame("BROWSE", bw)

	s := m
	s.view = viewSearch
	s.search.SetValue("jazz")
	frame("SEARCH", s)

	a := m
	a.view = viewAdd
	a.addName.SetValue("My Stream")
	a.addURL.SetValue("https://example.com/stream")
	frame("ADD STATION", a)

	st := m
	st.view = viewSettings
	frame("SETTINGS", st)

	f := m
	f.view = viewFavorites
	frame("FAVORITES", f)
}

func TestUpdateBannerText(t *testing.T) {
	mk := func(s update.Status) Model {
		return Model{st: newStyles(themeByName("catppuccin-mocha")), width: 80, update: s}
	}
	cases := []struct{ kind, want string }{
		{"homebrew", "brew upgrade"},
		{"scoop", "scoop update"},
		{"unknown", "releases"},
	}
	for _, c := range cases {
		m := mk(update.Status{Available: true, Latest: "v2.0.0", InstallKind: c.kind})
		if got := m.updateBanner(); !strings.Contains(got, c.want) {
			t.Errorf("kind %q banner = %q, want contains %q", c.kind, got, c.want)
		}
	}
	if got := mk(update.Status{Available: true, Latest: "v2.0.0", InstallKind: "binary", SelfUpdatable: true}).updateBanner(); !strings.Contains(got, "press u") {
		t.Errorf("self-updatable banner = %q, want key hint", got)
	}
	if got := (Model{}).updateBanner(); got != "" {
		t.Errorf("no update should render empty banner, got %q", got)
	}
	// Dismissed banner renders empty even when an update is available.
	dismissed := mk(update.Status{Available: true, Latest: "v2.0.0", InstallKind: "homebrew"})
	dismissed.updateDismiss = true
	if got := dismissed.updateBanner(); got != "" {
		t.Errorf("dismissed banner should be empty, got %q", got)
	}
}

func repeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}
