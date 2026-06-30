package tui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pedrosousa13/onda/internal/domain"
	"github.com/pedrosousa13/onda/internal/update"
)

var errSample = errors.New("boom")

func TestUpdateStationsMsgPopulatesList(t *testing.T) {
	m := Model{}
	updated, _ := m.Update(stationsMsg{stations: []domain.Station{{Name: "KEXP"}}})
	got := updated.(Model)
	if len(got.stations) != 1 || got.stations[0].Name != "KEXP" {
		t.Fatalf("stationsMsg did not populate list: %+v", got.stations)
	}
}

func TestUpdateErrMsgSetsStatus(t *testing.T) {
	m := Model{}
	updated, _ := m.Update(errMsg{err: errSample})
	if got := updated.(Model); got.status == "" {
		t.Fatal("errMsg should set a non-empty status line")
	}
}

type fakeStore struct {
	adds, removes int
	isFav         bool
	custom        []domain.Station
}

func (f *fakeStore) Favorites() ([]domain.Station, error)    { return nil, nil }
func (f *fakeStore) AddFavorite(domain.Station) error        { f.adds++; f.isFav = true; return nil }
func (f *fakeStore) RemoveFavorite(domain.Station) error     { f.removes++; f.isFav = false; return nil }
func (f *fakeStore) IsFavorite(domain.Station) (bool, error) { return f.isFav, nil }
func (f *fakeStore) AddCustom(s domain.Station) error        { f.custom = append(f.custom, s); return nil }
func (f *fakeStore) SaveQuality(domain.QualityPref) error    { return nil }
func (f *fakeStore) SaveTracking(string) error               { return nil }
func (f *fakeStore) SaveHistory(bool) error                  { return nil }
func (f *fakeStore) SaveTheme(string) error                  { return nil }
func (f *fakeStore) SaveUpdateCheck(bool) error              { return nil }

func TestToggleFavoriteAddsAndRemoves(t *testing.T) {
	fs := &fakeStore{}
	m := Model{store: fs, stations: []domain.Station{{Name: "KEXP", Homepage: "kexp.org"}}}
	updated, _ := m.toggleFavorite() // add
	m = updated
	updated, _ = m.toggleFavorite() // remove
	m = updated
	if fs.adds != 1 || fs.removes != 1 {
		t.Fatalf("toggle should add then remove, got adds=%d removes=%d", fs.adds, fs.removes)
	}
}

func TestNewStartsOnHome(t *testing.T) {
	m := New(nil, nil, nil, domain.QualityHighest, "never", false, "catppuccin-mocha", true, "1.0.0", t.TempDir())
	if m.view != viewHome {
		t.Fatalf("New should start on Home, got view %d", m.view)
	}
}

func TestUpdateMsgSetsBannerState(t *testing.T) {
	m := Model{}
	out, _ := m.Update(updateMsg{status: update.Status{
		Available: true, Latest: "v2.0.0", InstallKind: "homebrew",
	}})
	got := out.(Model)
	if !got.update.Available || got.update.Latest != "v2.0.0" {
		t.Fatalf("update status not stored: %+v", got.update)
	}
}

func TestUpdateKeyTriggersApplyOnlyWhenSelfUpdatable(t *testing.T) {
	// Not self-updatable: "u" must not start applying.
	m := Model{update: update.Status{Available: true, SelfUpdatable: false}}
	out, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("u")})
	if out.(Model).updateApplying {
		t.Fatal("u should not apply when not self-updatable")
	}
	// Self-updatable: "u" flips updateApplying and returns a command.
	m = Model{update: update.Status{Available: true, SelfUpdatable: true}}
	out, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("u")})
	if !out.(Model).updateApplying || cmd == nil {
		t.Fatal("u should start applying when self-updatable")
	}
}

func TestGoHomeSetsHomeView(t *testing.T) {
	m := Model{view: viewSearch, store: &fakeStore{}}
	updated, _ := m.goHome()
	if got := updated.(Model); got.view != viewHome {
		t.Fatalf("goHome should switch to Home, got view %d", got.view)
	}
}

func TestSettingsCycleQuality(t *testing.T) {
	m := Model{quality: domain.QualityHighest}
	m = m.cycleQuality()
	if m.quality == domain.QualityHighest {
		t.Fatal("cycleQuality should change the value")
	}
}
