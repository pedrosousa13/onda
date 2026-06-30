package tui

import (
	"errors"
	"fmt"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pedrosousa13/onda/internal/domain"
	"github.com/pedrosousa13/onda/internal/update"
)

type fakePlayer struct{ played string }

func (f *fakePlayer) Play(url string) error { f.played = url; return nil }
func (f *fakePlayer) Stop() error           { return nil }
func (f *fakePlayer) Volume(int) error      { return nil }

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

func TestReconnectSchedulesOnDrop(t *testing.T) {
	st := domain.Station{Name: "X", Variants: []domain.StreamVariant{{URL: "u", Bitrate: 128}}}
	m := Model{phase: phasePlaying, isPlaying: true, playing: st, varIdx: 0, player: &fakePlayer{}}
	out, cmd := m.Update(endedMsg{})
	got := out.(Model)
	if got.phase != phaseReconnecting || got.reconnectTry != 1 || cmd == nil {
		t.Fatalf("drop while playing should schedule a reconnect, got phase=%d try=%d hasCmd=%v",
			got.phase, got.reconnectTry, cmd != nil)
	}
}

func TestReconnectIgnoresEndWhenNotPlaying(t *testing.T) {
	// The previous file ending as we connect to a new station is not a drop.
	st := domain.Station{Name: "X", Variants: []domain.StreamVariant{{URL: "u"}}}
	m := Model{phase: phaseConnecting, playing: st, varIdx: 0, player: &fakePlayer{}}
	if got := mustModel(m.Update(endedMsg{})); got.phase != phaseConnecting {
		t.Fatalf("ended while connecting should be ignored, got phase=%d", got.phase)
	}
}

func TestReconnectGivesUpAfterMax(t *testing.T) {
	st := domain.Station{Name: "X", Variants: []domain.StreamVariant{{URL: "u"}}}
	m := Model{phase: phasePlaying, isPlaying: true, playing: st, varIdx: 0,
		reconnectTry: maxReconnect, player: &fakePlayer{}}
	got := mustModel(m.Update(endedMsg{}))
	if got.phase != phaseFailed || got.playErr == "" {
		t.Fatalf("should give up after max retries, got phase=%d err=%q", got.phase, got.playErr)
	}
}

func TestReconnectTickReplaysCurrentURL(t *testing.T) {
	fp := &fakePlayer{}
	st := domain.Station{Name: "X", Variants: []domain.StreamVariant{{URL: "u", Bitrate: 128}}}
	m := Model{phase: phaseReconnecting, playing: st, varIdx: 0, reconnectSeq: 3, player: fp}
	got := mustModel(m.Update(reconnectMsg{seq: 3}))
	if fp.played != "u" || got.phase != phaseConnecting {
		t.Fatalf("reconnect tick should replay and reconnect, got played=%q phase=%d", fp.played, got.phase)
	}
}

func TestReconnectStaleTickIgnored(t *testing.T) {
	fp := &fakePlayer{}
	st := domain.Station{Name: "X", Variants: []domain.StreamVariant{{URL: "u"}}}
	m := Model{phase: phaseReconnecting, playing: st, varIdx: 0, reconnectSeq: 5, player: fp}
	got := mustModel(m.Update(reconnectMsg{seq: 2})) // older generation
	if fp.played != "" || got.phase != phaseReconnecting {
		t.Fatalf("stale reconnect tick should be ignored, got played=%q phase=%d", fp.played, got.phase)
	}
}

func TestPlaySelectedResetsReconnectBudget(t *testing.T) {
	st := domain.Station{Name: "X", Variants: []domain.StreamVariant{{URL: "u", Bitrate: 128}}}
	m := Model{view: viewBrowse, stations: []domain.Station{st}, player: &fakePlayer{},
		quality: domain.QualityHighest, reconnectTry: 4}
	if got := mustModel(m.playSelected()); got.reconnectTry != 0 {
		t.Fatalf("playSelected should reset the reconnect budget, got %d", got.reconnectTry)
	}
}

func TestBufferingFromPlaying(t *testing.T) {
	m := Model{phase: phasePlaying, isPlaying: true}
	if got := mustModel(m.Update(bufferingMsg{})); got.phase != phaseBuffering {
		t.Fatalf("buffering while playing should enter phaseBuffering, got %d", got.phase)
	}
}

func TestBufferingIgnoredWhenConnecting(t *testing.T) {
	m := Model{phase: phaseConnecting}
	if got := mustModel(m.Update(bufferingMsg{})); got.phase != phaseConnecting {
		t.Fatalf("buffering while connecting should be ignored, got %d", got.phase)
	}
}

func TestIdleDuringBufferingDoesNotStop(t *testing.T) {
	m := Model{phase: phaseBuffering, isPlaying: true}
	got := mustModel(m.Update(idleMsg{}))
	if got.phase != phaseBuffering || !got.isPlaying {
		t.Fatalf("idle during buffering should not stop, got phase=%d playing=%v", got.phase, got.isPlaying)
	}
}

func TestPlayingClearsBuffering(t *testing.T) {
	m := Model{phase: phaseBuffering}
	got := mustModel(m.Update(playingMsg{}))
	if got.phase != phasePlaying || !got.isPlaying {
		t.Fatalf("playing should clear buffering, got phase=%d playing=%v", got.phase, got.isPlaying)
	}
}

func TestEndedReconnectsFromBuffering(t *testing.T) {
	st := domain.Station{Name: "X", Variants: []domain.StreamVariant{{URL: "u", Bitrate: 128}}}
	m := Model{phase: phaseBuffering, playing: st, varIdx: 0, player: &fakePlayer{}}
	if got := mustModel(m.Update(endedMsg{})); got.phase != phaseReconnecting {
		t.Fatalf("a drop while buffering should reconnect, got phase=%d", got.phase)
	}
}

func mustModel(model tea.Model, _ tea.Cmd) Model { return model.(Model) }

func searchModel(query string, seq int) Model {
	ti := textinput.New()
	ti.SetValue(query)
	return Model{view: viewSearch, search: ti, searchSeq: seq}
}

func TestSearchDebounceFiresLatestOnly(t *testing.T) {
	m := searchModel("jazz", 7)
	// A stale tick (older seq) must not trigger a search.
	if got := mustModel(m.Update(searchDebounceMsg{seq: 3})); got.loading {
		t.Fatal("stale debounce should not search")
	}
	// The latest tick with a long-enough query starts searching.
	if got := mustModel(m.Update(searchDebounceMsg{seq: 7})); !got.loading {
		t.Fatal("matching debounce should start searching")
	}
}

func TestSearchDebounceMinLength(t *testing.T) {
	m := searchModel("j", 1)
	if got := mustModel(m.Update(searchDebounceMsg{seq: 1})); got.loading {
		t.Fatal("query shorter than the minimum should not search")
	}
}

func TestSearchArrowNavigatesResults(t *testing.T) {
	m := searchModel("kexp", 1)
	m.stations = make([]domain.Station, 3)
	if got := mustModel(m.updateSearch(tea.KeyMsg{Type: tea.KeyDown})); got.cursor != 1 {
		t.Fatalf("down should move cursor to 1, got %d", got.cursor)
	}
}

func TestSearchEnterPlaysSelected(t *testing.T) {
	m := searchModel("kexp", 1)
	m.player = &fakePlayer{}
	m.quality = domain.QualityHighest
	m.stations = []domain.Station{
		{Name: "a", Variants: []domain.StreamVariant{{URL: "u", Bitrate: 128}}},
		{Name: "b", Variants: []domain.StreamVariant{{URL: "v", Bitrate: 128}}},
	}
	m.cursor = 1
	got := mustModel(m.updateSearch(tea.KeyMsg{Type: tea.KeyEnter}))
	if !got.isPlaying || got.view != viewBrowse {
		t.Fatalf("enter should play the selected result and open browse, got isPlaying=%v view=%d", got.isPlaying, got.view)
	}
}

func TestMouseWheelMovesCursor(t *testing.T) {
	m := Model{view: viewBrowse, stations: []domain.Station{{Name: "a"}, {Name: "b"}, {Name: "c"}}}
	if got := mustModel(m.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown})); got.cursor != 1 {
		t.Fatalf("wheel down should move cursor to 1, got %d", got.cursor)
	}
	m.cursor = 2 // last
	if got := mustModel(m.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown})); got.cursor != 2 {
		t.Fatalf("wheel down at end should stay at 2, got %d", got.cursor)
	}
}

func TestMouseClickSelectsThenPlays(t *testing.T) {
	stations := make([]domain.Station, 5)
	for i := range stations {
		stations[i] = domain.Station{Name: fmt.Sprintf("s%d", i), Variants: []domain.StreamVariant{{URL: "u", Bitrate: 128}}}
	}
	m := Model{view: viewBrowse, height: 20, stations: stations, player: &fakePlayer{}, quality: domain.QualityHighest}
	// Y=5 in a list view (rowStartY=3) → station index 2.
	got := mustModel(m.Update(tea.MouseMsg{Button: tea.MouseButtonLeft, Action: tea.MouseActionPress, Y: 5}))
	if got.cursor != 2 {
		t.Fatalf("click at Y=5 should select station 2, got cursor %d", got.cursor)
	}
	// Clicking the already-selected row plays it.
	played := mustModel(got.Update(tea.MouseMsg{Button: tea.MouseButtonLeft, Action: tea.MouseActionPress, Y: 5}))
	if !played.isPlaying || played.phase != phaseConnecting {
		t.Fatalf("second click should start playing, got isPlaying=%v phase=%d", played.isPlaying, played.phase)
	}
}

func TestReplaySameStationKeepsVariant(t *testing.T) {
	st := domain.Station{Name: "FIP", Homepage: "fip.fr", Variants: []domain.StreamVariant{
		{URL: "hi", Bitrate: 192}, {URL: "lo", Bitrate: 64},
	}}
	m := Model{
		view: viewBrowse, stations: []domain.Station{st},
		player: &fakePlayer{}, quality: domain.QualityHighest,
		isPlaying: true, playing: st, varIdx: 1, // user chose the 64k variant
	}
	if got := mustModel(m.playSelected()); got.varIdx != 1 {
		t.Fatalf("replaying the current station should keep varIdx 1 (64k), got %d", got.varIdx)
	}
}

func TestPlayDifferentStationUsesPreference(t *testing.T) {
	cur := domain.Station{Name: "FIP", Homepage: "fip.fr", Variants: []domain.StreamVariant{{URL: "x", Bitrate: 64}}}
	other := domain.Station{Name: "KEXP", Homepage: "kexp.org", Variants: []domain.StreamVariant{
		{URL: "hi", Bitrate: 192}, {URL: "lo", Bitrate: 64},
	}}
	m := Model{
		view: viewBrowse, stations: []domain.Station{other},
		player: &fakePlayer{}, quality: domain.QualityHighest,
		isPlaying: true, playing: cur, varIdx: 0,
	}
	// cursor 0 = other (KEXP); highest preference → 192k at index 0.
	if got := mustModel(m.playSelected()); got.varIdx != 0 || got.playing.Name != "KEXP" {
		t.Fatalf("different station should use preference, got varIdx=%d playing=%s", got.varIdx, got.playing.Name)
	}
}

func TestSettingsCycleQuality(t *testing.T) {
	m := Model{quality: domain.QualityHighest}
	m = m.cycleQuality()
	if m.quality == domain.QualityHighest {
		t.Fatal("cycleQuality should change the value")
	}
}
