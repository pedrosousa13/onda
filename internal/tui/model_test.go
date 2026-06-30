package tui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pedrosousa13/onda/internal/domain"
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
	m := New(nil, nil, nil, domain.QualityHighest, "never", false, "catppuccin-mocha")
	if m.view != viewHome {
		t.Fatalf("New should start on Home, got view %d", m.view)
	}
}

func TestGoHomeSetsHomeView(t *testing.T) {
	m := Model{view: viewSearch, store: &fakeStore{}}
	updated, _ := m.goHome()
	if got := updated.(Model); got.view != viewHome {
		t.Fatalf("goHome should switch to Home, got view %d", got.view)
	}
}

func TestPlaybackPhaseTransitions(t *testing.T) {
	// playing event → playing
	m := Model{phase: phaseConnecting, isPlaying: true}
	if got := mustModel(m.Update(playingMsg{})); got.phase != phasePlaying {
		t.Fatalf("playingMsg should set phasePlaying, got %d", got.phase)
	}
	// idle while playing → idle + stopped
	m = Model{phase: phasePlaying, isPlaying: true}
	if got := mustModel(m.Update(idleMsg{})); got.phase != phaseIdle || got.isPlaying {
		t.Fatalf("idleMsg while playing should idle+stop, got phase=%d isPlaying=%v", got.phase, got.isPlaying)
	}
	// idle while connecting → ignored (transient idle before playback)
	m = Model{phase: phaseConnecting, isPlaying: true}
	if got := mustModel(m.Update(idleMsg{})); got.phase != phaseConnecting {
		t.Fatalf("idleMsg while connecting should be ignored, got phase=%d", got.phase)
	}
	// error → failed with a message
	m = Model{phase: phaseConnecting, isPlaying: true}
	if got := mustModel(m.Update(playErrMsg{err: errSample})); got.phase != phaseFailed || got.playErr == "" {
		t.Fatalf("playErrMsg should fail with message, got phase=%d err=%q", got.phase, got.playErr)
	}
}

func TestConnectTimeoutGuard(t *testing.T) {
	// matching attempt while still connecting → failed
	m := Model{phase: phaseConnecting, playAttempt: 3}
	if got := mustModel(m.Update(connectTimeoutMsg{attempt: 3})); got.phase != phaseFailed {
		t.Fatalf("matching timeout should fail, got phase=%d", got.phase)
	}
	// stale attempt → ignored
	m = Model{phase: phaseConnecting, playAttempt: 5}
	if got := mustModel(m.Update(connectTimeoutMsg{attempt: 3})); got.phase != phaseConnecting {
		t.Fatalf("stale timeout should be ignored, got phase=%d", got.phase)
	}
}

func mustModel(model tea.Model, _ tea.Cmd) Model { return model.(Model) }

func TestSettingsCycleQuality(t *testing.T) {
	m := Model{quality: domain.QualityHighest}
	m = m.cycleQuality()
	if m.quality == domain.QualityHighest {
		t.Fatal("cycleQuality should change the value")
	}
}
