package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pedrosousa13/onda/internal/domain"
	"github.com/pedrosousa13/onda/internal/update"
)

type fakePlayer struct {
	played    string
	normalize bool
}

func (f *fakePlayer) Play(url string) error      { f.played = url; return nil }
func (f *fakePlayer) Stop() error                { return nil }
func (f *fakePlayer) Volume(int) error           { return nil }
func (f *fakePlayer) SetNormalize(on bool) error { f.normalize = on; return nil }

// stubDir satisfies Searcher with no-op local/network calls.
type stubDir struct{}

func (stubDir) Search(context.Context, string) ([]domain.Station, error) { return nil, nil }
func (stubDir) Popular(context.Context) ([]domain.Station, error)        { return nil, nil }
func (stubDir) Initial(context.Context) ([]domain.Station, error)        { return nil, nil }
func (stubDir) Refresh(context.Context) ([]domain.Station, error)        { return nil, nil }

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
	adds, removes  int
	isFav          bool
	custom         []domain.Station
	savedVolume    int
	savedNormalize bool
	recents        []domain.Station
	favs           []domain.Station
}

func (f *fakeStore) Favorites() ([]domain.Station, error)    { return f.favs, nil }
func (f *fakeStore) AddFavorite(domain.Station) error        { f.adds++; f.isFav = true; return nil }
func (f *fakeStore) RemoveFavorite(domain.Station) error     { f.removes++; f.isFav = false; return nil }
func (f *fakeStore) IsFavorite(domain.Station) (bool, error) { return f.isFav, nil }
func (f *fakeStore) AddCustom(s domain.Station) error        { f.custom = append(f.custom, s); return nil }
func (f *fakeStore) SaveQuality(domain.QualityPref) error    { return nil }
func (f *fakeStore) SaveTracking(string) error               { return nil }
func (f *fakeStore) SaveHistory(bool) error                  { return nil }
func (f *fakeStore) SaveTheme(string) error                  { return nil }
func (f *fakeStore) SaveUpdateCheck(bool) error              { return nil }
func (f *fakeStore) SaveLiveSearch(bool) error               { return nil }
func (f *fakeStore) SaveVolume(v int) error                  { f.savedVolume = v; return nil }
func (f *fakeStore) SaveNormalize(v bool) error              { f.savedNormalize = v; return nil }
func (f *fakeStore) Recents() ([]domain.Station, error)      { return f.recents, nil }
func (f *fakeStore) AddRecent(s domain.Station) error        { f.recents = append(f.recents, s); return nil }
func (f *fakeStore) ClearRecents() error                     { f.recents = nil; return nil }

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
	m := New(nil, nil, nil, domain.QualityHighest, "never", false, "catppuccin-mocha", true, true, 100, false, false, "1.0.0", t.TempDir())
	if m.view != viewHome {
		t.Fatalf("New should start on Home, got view %d", m.view)
	}
}

func TestSettingsToggleNormalize(t *testing.T) {
	fp := &fakePlayer{}
	fs := &fakeStore{}
	m := Model{view: viewSettings, player: fp, store: fs}
	got := mustModel(m.updateSettings(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("7")}))
	if !got.normalize || !fp.normalize || !fs.savedNormalize {
		t.Fatalf("toggling 7 should enable normalize everywhere, got model=%v player=%v store=%v",
			got.normalize, fp.normalize, fs.savedNormalize)
	}
}

func TestNewRestoresVolume(t *testing.T) {
	m := New(nil, nil, nil, domain.QualityHighest, "never", false, "catppuccin-mocha", true, true, 42, false, false, "1.0.0", t.TempDir())
	if m.volume != 42 {
		t.Fatalf("New should restore the saved volume, got %d", m.volume)
	}
	// Out-of-range values from a hand-edited config are clamped.
	m = New(nil, nil, nil, domain.QualityHighest, "never", false, "catppuccin-mocha", true, true, 150, false, false, "1.0.0", t.TempDir())
	if m.volume != 100 {
		t.Fatalf("New should clamp volume to 100, got %d", m.volume)
	}
}

func TestChangeVolumePersistsAndClamps(t *testing.T) {
	fs := &fakeStore{}
	m := Model{store: fs, player: &fakePlayer{}, volume: 98}
	got := mustModel(m.changeVolume(5)) // 98+5 = 103 → clamp to 100
	if got.volume != 100 {
		t.Fatalf("volume should clamp to 100, got %d", got.volume)
	}
	if fs.savedVolume != 100 {
		t.Fatalf("volume should be persisted, got %d", fs.savedVolume)
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

func mustModel(model tea.Model, _ tea.Cmd) Model { return model.(Model) }

func searchModel(query string, seq int) Model {
	ti := textinput.New()
	ti.SetValue(query)
	return Model{view: viewSearch, search: ti, searchSeq: seq, liveSearch: true}
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

func TestLiveSearchOffSkipsDebounce(t *testing.T) {
	m := searchModel("jaz", 0)
	m.liveSearch = false
	got := mustModel(m.updateSearch(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("z")}))
	if got.searchSeq != 0 {
		t.Fatalf("enter-to-search mode must not schedule a live search, got searchSeq=%d", got.searchSeq)
	}
}

func TestLiveSearchOffEnterSearchesNotPlays(t *testing.T) {
	m := searchModel("kexp", 0)
	m.liveSearch = false
	m.player = &fakePlayer{}
	m.quality = domain.QualityHighest
	// Stale stations from the prior view must not be played on enter.
	m.stations = []domain.Station{{Name: "stale", Variants: []domain.StreamVariant{{URL: "u", Bitrate: 128}}}}
	got := mustModel(m.updateSearch(tea.KeyMsg{Type: tea.KeyEnter}))
	if got.isPlaying {
		t.Fatal("enter-to-search mode must search, not play a stale station")
	}
	if !got.loading || got.view != viewBrowse {
		t.Fatalf("enter should start a search and open browse, got loading=%v view=%d", got.loading, got.view)
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

func TestPlayRecordsRecentOnlyWhenHistoryOn(t *testing.T) {
	st := domain.Station{Name: "KEXP", Homepage: "kexp.org", Variants: []domain.StreamVariant{{URL: "u", Bitrate: 128}}}

	off := &fakeStore{}
	m := Model{store: off, player: &fakePlayer{}, quality: domain.QualityHighest,
		stations: []domain.Station{st}, history: false}
	_ = mustModel(m.playSelected())
	if len(off.recents) != 0 {
		t.Fatalf("history off should record nothing, got %d", len(off.recents))
	}

	on := &fakeStore{}
	m = Model{store: on, player: &fakePlayer{}, quality: domain.QualityHighest,
		stations: []domain.Station{st}, history: true}
	_ = mustModel(m.playSelected())
	if len(on.recents) != 1 || on.recents[0].Name != "KEXP" {
		t.Fatalf("history on should record the played station, got %+v", on.recents)
	}
}

func TestShowRecentsLoadsList(t *testing.T) {
	fs := &fakeStore{recents: []domain.Station{{Name: "KEXP", Homepage: "kexp.org"}}}
	got := mustModel((Model{store: fs}).showRecents())
	if got.view != viewRecents || len(got.stations) != 1 {
		t.Fatalf("showRecents should open recents with items, got view=%d n=%d", got.view, len(got.stations))
	}
}

func TestClearRecentsEmptiesView(t *testing.T) {
	fs := &fakeStore{recents: []domain.Station{{Name: "KEXP"}}}
	m := Model{store: fs, view: viewRecents, stations: fs.recents}
	got := mustModel(m.clearRecents())
	if len(got.stations) != 0 || len(fs.recents) != 0 {
		t.Fatalf("clear should empty recents, got view=%d store=%d", len(got.stations), len(fs.recents))
	}
}

func TestHomeSeedsRecentsAboveFavorites(t *testing.T) {
	fs := &fakeStore{
		recents: []domain.Station{{Name: "KEXP", Homepage: "kexp.org"}, {Name: "FIP", Homepage: "fip.fr"}},
		favs:    []domain.Station{{Name: "NTS", Homepage: "nts.live"}},
	}
	m := New(nil, nil, fs, domain.QualityHighest, "never", true, "catppuccin-mocha", true, true, 100, false, false, "1.0.0", t.TempDir())
	if got := m.homeRecentsN(); got != 2 {
		t.Fatalf("homeRecentsN = %d, want 2", got)
	}
	if len(m.stations) != 3 {
		t.Fatalf("stations = %d, want 3 (2 recents + 1 favorite)", len(m.stations))
	}
	if m.stations[0].Name != "KEXP" || m.stations[2].Name != "NTS" {
		t.Fatalf("recents should lead, favorites follow, got %s…%s", m.stations[0].Name, m.stations[2].Name)
	}
}

func TestHomeNoRecentsWhenHistoryOff(t *testing.T) {
	fs := &fakeStore{
		recents: []domain.Station{{Name: "KEXP", Homepage: "kexp.org"}},
		favs:    []domain.Station{{Name: "NTS", Homepage: "nts.live"}},
	}
	m := New(nil, nil, fs, domain.QualityHighest, "never", false, "catppuccin-mocha", true, true, 100, false, false, "1.0.0", t.TempDir())
	if got := m.homeRecentsN(); got != 0 {
		t.Fatalf("history off → homeRecentsN = %d, want 0", got)
	}
	if len(m.stations) != 1 || m.stations[0].Name != "NTS" {
		t.Fatalf("history off → home should show favorites only, got %+v", m.stations)
	}
}

func TestHomeRecentsCappedAtFive(t *testing.T) {
	var rec []domain.Station
	for i := 0; i < 7; i++ {
		rec = append(rec, domain.Station{Name: fmt.Sprintf("S%d", i), Homepage: fmt.Sprintf("s%d", i)})
	}
	fs := &fakeStore{recents: rec, favs: []domain.Station{{Name: "NTS"}}}
	m := New(nil, nil, fs, domain.QualityHighest, "never", true, "catppuccin-mocha", true, true, 100, false, false, "1.0.0", t.TempDir())
	if got := m.homeRecentsN(); got != homeRecentsCap {
		t.Fatalf("home recents = %d, want cap %d", got, homeRecentsCap)
	}
}

func TestHomeRendersRecentLabel(t *testing.T) {
	fs := &fakeStore{
		recents: []domain.Station{{Name: "KEXP", Homepage: "kexp.org"}},
		favs:    []domain.Station{{Name: "NTS", Homepage: "nts.live"}},
	}
	on := New(nil, nil, fs, domain.QualityHighest, "never", true, "catppuccin-mocha", true, true, 100, false, false, "1.0.0", t.TempDir())
	on.width, on.height = 76, 24
	if !strings.Contains(on.View(), "recent") {
		t.Fatal("home with history on should render a 'recent' section label")
	}

	off := New(nil, nil, fs, domain.QualityHighest, "never", false, "catppuccin-mocha", true, true, 100, false, false, "1.0.0", t.TempDir())
	off.width, off.height = 76, 24
	if strings.Contains(off.View(), "recent") {
		t.Fatal("home with history off should not render a 'recent' section label")
	}
}

func TestHomeStationAtYMapsBothSections(t *testing.T) {
	fs := &fakeStore{
		recents: []domain.Station{{Name: "KEXP", Homepage: "kexp.org"}, {Name: "FIP", Homepage: "fip.fr"}},
		favs:    []domain.Station{{Name: "NTS", Homepage: "nts.live"}},
	}
	m := New(nil, nil, fs, domain.QualityHighest, "never", true, "catppuccin-mocha", true, true, 100, false, false, "1.0.0", t.TempDir())
	m.width, m.height = 76, 24
	// Layout: "recent" label at y=10, recents rows at y=11,12, "favorites" at y=13, first fav at y=14.
	if got := m.stationAtY(11); got != 0 {
		t.Fatalf("y=11 should be first recent (idx 0), got %d", got)
	}
	if got := m.stationAtY(14); got != 2 {
		t.Fatalf("y=14 should be first favorite (idx 2), got %d", got)
	}
}

func TestSettingsCycleQuality(t *testing.T) {
	m := Model{quality: domain.QualityHighest}
	m = m.cycleQuality()
	if m.quality == domain.QualityHighest {
		t.Fatal("cycleQuality should change the value")
	}
}

func TestRKeyTriggersRefresh(t *testing.T) {
	m := Model{view: viewBrowse, dir: stubDir{}}
	out, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("R")})
	got := out.(Model)
	if !got.refreshing || cmd == nil {
		t.Fatalf("R should start a refresh: refreshing=%v cmd=%v", got.refreshing, cmd)
	}
}

func TestCorpusRefreshedClearsRefreshingAndRepullsPreview(t *testing.T) {
	m := Model{view: viewBrowse, crumb: "popular", refreshing: true, dir: stubDir{}}
	out, cmd := m.Update(corpusRefreshedMsg{stations: make([]domain.Station, 10)})
	got := out.(Model)
	if got.refreshing {
		t.Fatal("refresh complete should clear the refreshing indicator")
	}
	if cmd == nil {
		t.Fatal("on the popular preview, refresh-complete should re-pull the list")
	}
}

func TestCorpusRefreshedDoesNotClobberSearch(t *testing.T) {
	m := Model{view: viewBrowse, crumb: "“jazz”", refreshing: true,
		stations: make([]domain.Station, 4), dir: stubDir{}}
	got := mustModel(m.Update(corpusRefreshedMsg{stations: make([]domain.Station, 10)}))
	if got.refreshing {
		t.Fatal("refreshing should clear")
	}
	if len(got.stations) != 4 {
		t.Fatalf("search results must be preserved, got %d", len(got.stations))
	}
}
