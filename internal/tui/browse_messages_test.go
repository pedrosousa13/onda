package tui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pedrosousa13/onda/internal/domain"
)

// fakeDir implements the full Searcher interface with fixture data for testing.
type fakeDir struct{}

func (f fakeDir) Search(ctx context.Context, query string) ([]domain.Station, error) {
	return nil, nil
}

func (f fakeDir) Popular(ctx context.Context) ([]domain.Station, error) {
	return nil, nil
}

func (f fakeDir) Initial(ctx context.Context) ([]domain.Station, error) {
	return nil, nil
}

func (f fakeDir) Refresh(ctx context.Context) ([]domain.Station, error) {
	return nil, nil
}

func (f fakeDir) RefreshWithProgress(ctx context.Context, onProgress func(downloaded int64)) ([]domain.Station, error) {
	return nil, nil
}

func (f fakeDir) ClearCorpus() error {
	return nil
}

func (f fakeDir) CorpusSize() (int64, bool) {
	return 0, false
}

func (f fakeDir) Countries(ctx context.Context) ([]domain.Facet, error) {
	return []domain.Facet{
		{Name: "Portugal", Count: 10},
		{Name: "Brazil", Count: 8},
	}, nil
}

func (f fakeDir) Tags(ctx context.Context) ([]domain.Facet, error) {
	return []domain.Facet{
		{Name: "Rock", Count: 15},
		{Name: "Jazz", Count: 12},
	}, nil
}

func (f fakeDir) Languages(ctx context.Context) ([]domain.Facet, error) {
	return []domain.Facet{
		{Name: "Portuguese", Count: 20},
		{Name: "English", Count: 18},
	}, nil
}

func (f fakeDir) StationsBy(ctx context.Context, axis domain.Axis, value string, srt domain.Sort) ([]domain.Station, error) {
	return []domain.Station{
		{Name: "Station 1", Country: "Portugal", Votes: 100},
		{Name: "Station 2", Country: "Portugal", Votes: 50},
	}, nil
}

func TestFacetsCmdCountries(t *testing.T) {
	cmd := facetsCmd(fakeDir{}, domain.AxisCountry)
	if cmd == nil {
		t.Fatal("facetsCmd returned nil")
	}

	msg := cmd()
	facetsMsg, ok := msg.(facetsMsg)
	if !ok {
		t.Fatalf("expected facetsMsg, got %T", msg)
	}

	if facetsMsg.axis != domain.AxisCountry {
		t.Errorf("expected axis AxisCountry, got %v", facetsMsg.axis)
	}

	if len(facetsMsg.facets) == 0 {
		t.Error("expected non-empty facets")
	}

	// Verify it's countries data
	if facetsMsg.facets[0].Name != "Portugal" {
		t.Errorf("expected first country 'Portugal', got %s", facetsMsg.facets[0].Name)
	}
}

func TestFacetsCmdTags(t *testing.T) {
	cmd := facetsCmd(fakeDir{}, domain.AxisTag)
	if cmd == nil {
		t.Fatal("facetsCmd returned nil")
	}

	msg := cmd()
	facetsMsg, ok := msg.(facetsMsg)
	if !ok {
		t.Fatalf("expected facetsMsg, got %T", msg)
	}

	if facetsMsg.axis != domain.AxisTag {
		t.Errorf("expected axis AxisTag, got %v", facetsMsg.axis)
	}

	// Verify it's tags data (distinct from countries)
	if facetsMsg.facets[0].Name != "Rock" {
		t.Errorf("expected first tag 'Rock', got %s", facetsMsg.facets[0].Name)
	}
}

func TestFacetsCmdLanguages(t *testing.T) {
	cmd := facetsCmd(fakeDir{}, domain.AxisLanguage)
	if cmd == nil {
		t.Fatal("facetsCmd returned nil")
	}

	msg := cmd()
	facetsMsg, ok := msg.(facetsMsg)
	if !ok {
		t.Fatalf("expected facetsMsg, got %T", msg)
	}

	if facetsMsg.axis != domain.AxisLanguage {
		t.Errorf("expected axis AxisLanguage, got %v", facetsMsg.axis)
	}

	// Verify it's languages data (distinct from countries and tags)
	if facetsMsg.facets[0].Name != "Portuguese" {
		t.Errorf("expected first language 'Portuguese', got %s", facetsMsg.facets[0].Name)
	}
}

func TestStationsByCmdReturnsStationsMsg(t *testing.T) {
	cmd := stationsByCmd(fakeDir{}, domain.AxisCountry, "Portugal", domain.Sort{})
	if cmd == nil {
		t.Fatal("stationsByCmd returned nil")
	}

	msg := cmd()
	stationsMsg, ok := msg.(stationsMsg)
	if !ok {
		t.Fatalf("expected stationsMsg, got %T", msg)
	}

	if len(stationsMsg.stations) == 0 {
		t.Error("expected non-empty stations")
	}

	if stationsMsg.stations[0].Name != "Station 1" {
		t.Errorf("expected first station 'Station 1', got %s", stationsMsg.stations[0].Name)
	}
}

func TestFacetsCmdError(t *testing.T) {
	// Create a fake dir that returns an error
	badDir := &badFakeDirForError{}
	cmd := facetsCmd(badDir, domain.AxisCountry)
	if cmd == nil {
		t.Fatal("facetsCmd returned nil")
	}

	msg := cmd()
	errMsg, ok := msg.(errMsg)
	if !ok {
		t.Fatalf("expected errMsg, got %T", msg)
	}

	if errMsg.err == nil {
		t.Error("expected non-nil error")
	}
}

func TestStationsByCmdError(t *testing.T) {
	// Create a fake dir that returns an error
	badDir := &badFakeDirForError{}
	cmd := stationsByCmd(badDir, domain.AxisCountry, "Portugal", domain.Sort{})
	if cmd == nil {
		t.Fatal("stationsByCmd returned nil")
	}

	msg := cmd()
	errMsg, ok := msg.(errMsg)
	if !ok {
		t.Fatalf("expected errMsg, got %T", msg)
	}

	if errMsg.err == nil {
		t.Error("expected non-nil error")
	}
}

// badFakeDirForError implements Searcher but returns errors for browse methods
type badFakeDirForError struct{}

func (b *badFakeDirForError) Search(ctx context.Context, query string) ([]domain.Station, error) {
	return nil, nil
}

func (b *badFakeDirForError) Popular(ctx context.Context) ([]domain.Station, error) {
	return nil, nil
}

func (b *badFakeDirForError) Initial(ctx context.Context) ([]domain.Station, error) {
	return nil, nil
}

func (b *badFakeDirForError) Refresh(ctx context.Context) ([]domain.Station, error) {
	return nil, nil
}

func (b *badFakeDirForError) RefreshWithProgress(ctx context.Context, onProgress func(downloaded int64)) ([]domain.Station, error) {
	return nil, nil
}

func (b *badFakeDirForError) ClearCorpus() error {
	return nil
}

func (b *badFakeDirForError) CorpusSize() (int64, bool) {
	return 0, false
}

func (b *badFakeDirForError) Countries(ctx context.Context) ([]domain.Facet, error) {
	return nil, tea.ErrProgramKilled
}

func (b *badFakeDirForError) Tags(ctx context.Context) ([]domain.Facet, error) {
	return nil, tea.ErrProgramKilled
}

func (b *badFakeDirForError) Languages(ctx context.Context) ([]domain.Facet, error) {
	return nil, tea.ErrProgramKilled
}

func (b *badFakeDirForError) StationsBy(ctx context.Context, axis domain.Axis, value string, srt domain.Sort) ([]domain.Station, error) {
	return nil, tea.ErrProgramKilled
}
