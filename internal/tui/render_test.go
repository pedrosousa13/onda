package tui

import (
	"fmt"
	"testing"

	"github.com/pedrosousa13/radio/internal/domain"
)

// sampleModel builds a populated model for visual inspection of View() output.
func sampleModel() Model {
	m := New(nil, nil, nil, domain.QualityHighest, "never", false)
	m.stations = []domain.Station{
		{Name: "KEXP", Country: "United States", Tags: []string{"indie", "seattle"}},
		{Name: "BBC World Service", Country: "United Kingdom", Tags: []string{"news"}},
		{Name: "FIP", Country: "France", Tags: []string{"eclectic"}},
		{Name: "NTS Radio 1", Country: "United Kingdom", Tags: []string{"electronic"}},
		{Name: "Radio Nacional de España", Country: "Spain"},
		{Name: "Triple J", Country: "Australia", Tags: []string{"alternative"}},
	}
	m.cursor = 2
	m.nowTitle = "Khruangbin — Maria También"
	m.status = "playing: FIP"
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
	frame("BROWSE", m)

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

func repeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}
