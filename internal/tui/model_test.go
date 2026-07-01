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
func (stubDir) RefreshWithProgress(context.Context, func(int64)) ([]domain.Station, error) {
	return nil, nil
}
func (stubDir) ClearCorpus() error        { return nil }
func (stubDir) CorpusSize() (int64, bool) { return 0, false }
func (stubDir) Countries(context.Context) ([]domain.Facet, error) {
	return nil, nil
}
func (stubDir) Tags(context.Context) ([]domain.Facet, error) {
	return nil, nil
}
func (stubDir) Languages(context.Context) ([]domain.Facet, error) {
	return nil, nil
}
func (stubDir) StationsBy(context.Context, domain.Axis, string, domain.Sort) ([]domain.Station, error) {
	return nil, nil
}

var errSample = errors.New("boom")

func TestCorpusRefreshedKeepsFavoritesOnHome(t *testing.T) {
	favs := []domain.Station{{Name: "KEXP", Homepage: "kexp.org"}}
	m := Model{view: viewHome, store: &fakeStore{favs: favs}, dir: stubDir{}}
	m.markFavorites() // favKeys now non-empty (user has favorites)
	m.stations = favs
	out, _ := m.Update(corpusRefreshedMsg{stations: []domain.Station{{Name: "PopularOnly"}}})
	got := out.(Model)
	if got.loading {
		t.Fatal("refresh must not reload/clobber the favorites list on Home")
	}
	if len(got.stations) != 1 || got.stations[0].Name != "KEXP" {
		t.Fatalf("favorites should be preserved on Home, got %+v", got.stations)
	}
}

func TestCorpusRefreshedReloadsPopularWhenNoFavorites(t *testing.T) {
	m := Model{view: viewHome, store: &fakeStore{}, dir: stubDir{}} // no favorites
	m.markFavorites()
	out, cmd := m.Update(corpusRefreshedMsg{stations: nil})
	if !out.(Model).loading || cmd == nil {
		t.Fatal("with no favorites, Home should reload the vote-sorted popular preview")
	}
}

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
	savedCatalog   string
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
func (f *fakeStore) SaveOfflineCatalog(v string) error       { f.savedCatalog = v; return nil }
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
	m := New(nil, nil, nil, domain.QualityHighest, "never", false, "catppuccin-mocha", true, true, 100, false, false, "ask", "1.0.0", t.TempDir())
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

func TestDisableCatalogPersistsOff(t *testing.T) {
	fs := &fakeStore{}
	m := Model{offlineCatalog: "on", store: fs}
	m2 := m.disableCatalog()
	if m2.offlineCatalog != "off" || fs.savedCatalog != "off" {
		t.Fatalf("expected off+persisted, got %q / %q", m2.offlineCatalog, fs.savedCatalog)
	}
}

func TestClearCatalogCacheDisablesAndClears(t *testing.T) {
	fs := &fakeStore{}
	m := Model{view: viewSettings, offlineCatalog: "on", store: fs, dir: stubDir{}}
	got := mustModel(m.updateSettings(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("9")}))
	if got.offlineCatalog != "off" || fs.savedCatalog != "off" {
		t.Fatalf("clear should set off+persist, got %q / %q", got.offlineCatalog, fs.savedCatalog)
	}
	if got.status == "" {
		t.Fatal("expected a status message after clearing")
	}
}

func TestNewRestoresVolume(t *testing.T) {
	m := New(nil, nil, nil, domain.QualityHighest, "never", false, "catppuccin-mocha", true, true, 42, false, false, "ask", "1.0.0", t.TempDir())
	if m.volume != 42 {
		t.Fatalf("New should restore the saved volume, got %d", m.volume)
	}
	// Out-of-range values from a hand-edited config are clamped.
	m = New(nil, nil, nil, domain.QualityHighest, "never", false, "catppuccin-mocha", true, true, 150, false, false, "ask", "1.0.0", t.TempDir())
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
	m := New(nil, nil, fs, domain.QualityHighest, "never", true, "catppuccin-mocha", true, true, 100, false, false, "ask", "1.0.0", t.TempDir())
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
	m := New(nil, nil, fs, domain.QualityHighest, "never", false, "catppuccin-mocha", true, true, 100, false, false, "ask", "1.0.0", t.TempDir())
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
	m := New(nil, nil, fs, domain.QualityHighest, "never", true, "catppuccin-mocha", true, true, 100, false, false, "ask", "1.0.0", t.TempDir())
	if got := m.homeRecentsN(); got != homeRecentsCap {
		t.Fatalf("home recents = %d, want cap %d", got, homeRecentsCap)
	}
}

func TestHomeRendersRecentLabel(t *testing.T) {
	fs := &fakeStore{
		recents: []domain.Station{{Name: "KEXP", Homepage: "kexp.org"}},
		favs:    []domain.Station{{Name: "NTS", Homepage: "nts.live"}},
	}
	on := New(nil, nil, fs, domain.QualityHighest, "never", true, "catppuccin-mocha", true, true, 100, false, false, "ask", "1.0.0", t.TempDir())
	on.width, on.height = 76, 24
	if !strings.Contains(on.View(), "recent") {
		t.Fatal("home with history on should render a 'recent' section label")
	}

	off := New(nil, nil, fs, domain.QualityHighest, "never", false, "catppuccin-mocha", true, true, 100, false, false, "ask", "1.0.0", t.TempDir())
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
	m := New(nil, nil, fs, domain.QualityHighest, "never", true, "catppuccin-mocha", true, true, 100, false, false, "ask", "1.0.0", t.TempDir())
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

func TestInitDoesNotDownloadWhenConsentAsk(t *testing.T) {
	m := Model{offlineCatalog: "ask", needsRefresh: false} // app gates: ask => needsRefresh false
	_ = m.Init()
	if m.refreshing {
		t.Fatal("must not auto-download when consent is 'ask'")
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

func TestCorpusProgressUpdatesThenCompletes(t *testing.T) {
	m := Model{refreshing: true, progress: make(chan int64, 1)}
	m2, _ := m.Update(corpusProgressMsg{downloaded: 1024})
	if got := m2.(Model).downloaded; got != 1024 {
		t.Fatalf("downloaded = %d, want 1024", got)
	}
	m3, _ := m2.(Model).Update(corpusRefreshedMsg{stations: []domain.Station{{Name: "X"}}})
	if m3.(Model).refreshing {
		t.Fatal("refreshing should be false after completion")
	}
}

func TestHomeBannerEnableStartsDownload(t *testing.T) {
	fs := &fakeStore{}
	m := Model{view: viewHome, offlineCatalog: "ask", store: fs, dir: stubDir{}}
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	mm := m2.(Model)
	if mm.offlineCatalog != "on" {
		t.Fatalf("consent = %q, want on", mm.offlineCatalog)
	}
	if fs.savedCatalog != "on" {
		t.Fatal("consent not persisted")
	}
	if !mm.refreshing {
		t.Fatal("download not started")
	}
}

// keyMsg builds a tea.KeyMsg from a key name: "esc", "enter", or a single rune.
func keyMsg(k string) tea.KeyMsg {
	switch k {
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	default:
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
	}
}

func TestBrowseOpensMenu(t *testing.T) {
	m := Model{view: viewHome, dir: fakeDir{}}
	got := mustModel(m.openBrowse())
	if got.view != viewBrowseMenu {
		t.Fatalf("openBrowse should switch to viewBrowseMenu, got view %d", got.view)
	}
	if got.browseLevel != 0 {
		t.Fatalf("openBrowse should start at level 0, got %d", got.browseLevel)
	}
	if len(got.facets) != 3 {
		t.Fatalf("openBrowse should offer 3 axis facets, got %d", len(got.facets))
	}
}

func TestBrowseFacetsMsgAdvancesToLevel1(t *testing.T) {
	m := Model{view: viewBrowseMenu, dir: fakeDir{}}
	facets := []domain.Facet{{Name: "Portugal", Count: 10}, {Name: "Brazil", Count: 8}}
	got := mustModel(m.Update(facetsMsg{axis: domain.AxisCountry, facets: facets}))
	if got.browseLevel != 1 {
		t.Fatalf("facetsMsg should advance to level 1, got %d", got.browseLevel)
	}
	if got.browseAxis != domain.AxisCountry {
		t.Fatalf("facetsMsg should set browseAxis, got %v", got.browseAxis)
	}
	if len(got.facets) != 2 {
		t.Fatalf("facetsMsg should populate facets, got %d", len(got.facets))
	}
	if got.loading {
		t.Fatal("facetsMsg should clear loading")
	}
}

func TestBrowseEscPopsFromLevel2(t *testing.T) {
	m := Model{view: viewBrowse, dir: fakeDir{}, browseLevel: 2, browseAxis: domain.AxisCountry, browseValue: "Portugal"}
	out, _ := m.handleKey(keyMsg("esc"))
	got := out.(Model)
	if got.view != viewBrowseMenu {
		t.Fatalf("esc from browseLevel 2 should return to viewBrowseMenu, got view %d", got.view)
	}
	if got.browseLevel != 1 {
		t.Fatalf("esc from browseLevel 2 should drop to level 1, got %d", got.browseLevel)
	}
}

func TestBrowseSortCycleO(t *testing.T) {
	m := Model{view: viewBrowse, dir: fakeDir{}, browseLevel: 2, browseAxis: domain.AxisCountry, browseValue: "Portugal"}
	out, cmd := m.handleKey(keyMsg("o"))
	got := out.(Model)
	if got.browseSort.Key != domain.SortName {
		t.Fatalf("o should cycle sort to SortName, got %v", got.browseSort.Key)
	}
	if cmd == nil {
		t.Fatal("o should return a load command")
	}
}

func TestBrowseReverseO(t *testing.T) {
	m := Model{view: viewBrowse, dir: fakeDir{}, browseLevel: 2, browseAxis: domain.AxisCountry, browseValue: "Portugal"}
	out, cmd := m.handleKey(keyMsg("O"))
	got := out.(Model)
	if !got.browseSort.Flip {
		t.Fatal("O should flip browseSort.Flip to true")
	}
	if cmd == nil {
		t.Fatal("O should return a load command")
	}
}

func TestBrowseSortInertOutsideLevel2(t *testing.T) {
	m := Model{view: viewBrowse, dir: fakeDir{}, browseLevel: 0}
	out, _ := m.handleKey(keyMsg("o"))
	got := out.(Model)
	if got.browseSort != (domain.Sort{}) {
		t.Fatalf("o outside browseLevel 2 should be inert, got %+v", got.browseSort)
	}
	out, _ = got.handleKey(keyMsg("O"))
	got = out.(Model)
	if got.browseSort != (domain.Sort{}) {
		t.Fatalf("O outside browseLevel 2 should be inert, got %+v", got.browseSort)
	}
}

func TestShouldOfferCatalogHint(t *testing.T) {
	m := Model{offlineCatalog: "ask"}
	if !m.shouldOfferCatalogHint("raido eins", 0) {
		t.Fatal("want hint: ask + real query + zero results")
	}
	m.offlineCatalog = "on"
	if m.shouldOfferCatalogHint("raido eins", 0) {
		t.Fatal("no hint once catalog is on")
	}
	m.offlineCatalog = "ask"
	if m.shouldOfferCatalogHint("raido eins", 5) {
		t.Fatal("no hint when there are results")
	}
	if m.shouldOfferCatalogHint("r", 0) {
		t.Fatal("no hint for sub-minSearchLen queries")
	}
}
