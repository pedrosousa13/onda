package tui

import (
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pedrosousa13/onda/internal/domain"
	"github.com/pedrosousa13/onda/internal/update"
)

type view int

const (
	viewHome view = iota // landing: now-playing + favorites
	viewBrowse
	viewSearch
	viewFavorites
	viewAdd
	viewSettings
	viewRecents
)

// playbackPhase tracks what the now-playing panel should honestly report.
type playbackPhase int

const (
	phaseIdle playbackPhase = iota
	phaseConnecting
	phasePlaying
	phaseFailed
)

// connectTimeout marks a play attempt as failed if it never starts playing.
const connectTimeout = 12 * time.Second

// homeRecentsCap bounds the "recent" section shown above favorites on Home.
const homeRecentsCap = 5

// catalogSizeHint is the approximate download size shown in the offer UI.
// (Task 10 replaces this with a measured figure.)
const catalogSizeHint = "~30 MB"

type Player interface {
	Play(url string) error
	Stop() error
	Volume(pct int) error
	SetNormalize(on bool) error
}

// Store is the persistence slice the TUI needs.
type Store interface {
	Favorites() ([]domain.Station, error)
	AddFavorite(domain.Station) error
	RemoveFavorite(domain.Station) error
	IsFavorite(domain.Station) (bool, error)
	AddCustom(domain.Station) error
	SaveQuality(domain.QualityPref) error
	SaveTracking(string) error
	SaveHistory(bool) error
	SaveTheme(string) error
	SaveOfflineCatalog(string) error
	SaveUpdateCheck(bool) error
	SaveLiveSearch(bool) error
	SaveVolume(int) error
	SaveNormalize(bool) error
	Recents() ([]domain.Station, error)
	AddRecent(domain.Station) error
	ClearRecents() error
}

type Model struct {
	dir             Searcher
	player          Player
	store           Store
	view            view
	stations        []domain.Station
	cursor          int
	hoverIdx        int // station row under the mouse, -1 if none
	status          string
	nowTitle        string
	playing         domain.Station
	varIdx          int // index into playing.Variants currently streaming
	isPlaying       bool
	phase           playbackPhase
	playErr         string // message shown when phase == phaseFailed
	playAttempt     int    // monotonic; guards stale connect timeouts
	quality         domain.QualityPref
	tracking        string
	history         bool
	volume          int
	normalize       bool
	themeName       string
	st              Styles
	width           int
	height          int
	favKeys         map[string]bool
	homeRecents     []domain.Station // leading "recent" section on Home (opt-in, capped)
	sp              spinner.Model
	loading         bool
	refreshing      bool
	needsRefresh    bool
	offlineCatalog  string // consent: ask|on|off — gates auto-download in Init
	bannerDismissed bool   // "not now" on the Home offline-catalog banner this session
	downloaded      int64
	progress        chan int64
	crumb           string

	update         update.Status
	updateDismiss  bool
	updateApplying bool
	updateCheck    bool   // user preference (drives settings toggle)
	version        string // build version, for the update check
	updateCacheDir string // where update-cache.json lives

	search     textinput.Model
	searchSeq  int  // live-search debounce generation
	liveSearch bool // search as you type; off → enter-to-search
	addName    textinput.Model
	addURL     textinput.Model
	addBr      textinput.Model
	addFocus   int // 0=name, 1=url, 2=bitrate
}

func New(dir Searcher, p Player, st Store, quality domain.QualityPref, tracking string, history bool, theme string, updateCheck, liveSearch bool, volume int, normalize bool, needsRefresh bool, consent string, version, updateCacheDir string) Model {
	search := textinput.New()
	search.Placeholder = "search stations, country, or genre…"
	name := textinput.New()
	name.Placeholder = "name"
	url := textinput.New()
	url.Placeholder = "https://stream-url"
	br := textinput.New()
	br.Placeholder = "bitrate kbps (optional)"

	t := themeByName(theme)
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	m := Model{
		dir: dir, player: p, store: st,
		quality: quality, tracking: tracking, history: history,
		volume: clampVolume(volume), normalize: normalize, themeName: t.Name, st: newStyles(t),
		width: 80, height: 24, favKeys: map[string]bool{},
		hoverIdx: -1,
		sp:       sp, view: viewHome, crumb: "home",
		needsRefresh: needsRefresh, refreshing: needsRefresh, offlineCatalog: consent,
		updateCheck: updateCheck, version: version, updateCacheDir: updateCacheDir,
		liveSearch: liveSearch,
		search:     search, addName: name, addURL: url, addBr: br,
	}
	// Seed Home: opt-in recents on top, then favorites; if there are no
	// favorites, Init fetches a Popular preview (recents stay pinned on top).
	if st != nil {
		m.homeRecents = m.recentsForHome()
		favs, err := st.Favorites()
		if err != nil {
			favs = nil
		}
		m.stations = append(append([]domain.Station{}, m.homeRecents...), favs...)
		m.markFavorites()
		if len(favs) == 0 {
			m.loading = true
		}
	} else {
		m.loading = true
	}
	// Created here, not in Init(): Init has a value receiver, so it can't
	// persist a channel onto the running model. The R-key path uses startRefresh().
	if needsRefresh {
		m.progress = make(chan int64, 1)
	}
	return m
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{}
	if m.loading { // no favorites yet → load an initial preview for Home
		cmds = append(cmds, initialCmd(m.dir), m.sp.Tick)
	}
	if m.needsRefresh { // refresh the corpus in the background
		cmds = append(cmds, refreshWithProgressCmd(m.dir, m.progress), listenProgressCmd(m.progress), m.sp.Tick)
	}
	if m.updateCheck && m.version != "" && !strings.HasSuffix(m.version, "-dev") {
		cmds = append(cmds, updateCheckCmd(m.version, m.updateCacheDir))
	}
	if len(cmds) == 0 {
		return nil
	}
	return tea.Batch(cmds...)
}

// load starts an async station fetch and shows the spinner until it returns.
func (m Model) load(cmd tea.Cmd) (tea.Model, tea.Cmd) {
	m.loading = true
	return m, tea.Batch(cmd, m.sp.Tick)
}

// startRefresh kicks off a background corpus download with live byte progress.
// It no-ops if a download is already in flight, so rapid enable/toggle (banner,
// hint, settings, or R) can't spawn concurrent full-dump fetches.
func (m Model) startRefresh() (Model, tea.Cmd) {
	if m.refreshing {
		return m, nil
	}
	m.progress = make(chan int64, 1)
	m.refreshing = true
	return m, tea.Batch(refreshWithProgressCmd(m.dir, m.progress), listenProgressCmd(m.progress), m.sp.Tick)
}

// enableCatalog opts in, persists consent, and starts the background download.
// Shared by the Home banner, the search hint, and the settings row.
func (m Model) enableCatalog() (Model, tea.Cmd) {
	m.offlineCatalog = "on"
	if m.store != nil {
		_ = m.store.SaveOfflineCatalog("on")
	}
	return m.startRefresh()
}

// disableCatalog turns the offline catalog off and persists the choice; it does
// not delete any cached dump (the user can remove the file by hand).
func (m Model) disableCatalog() Model {
	m.offlineCatalog = "off"
	if m.store != nil {
		_ = m.store.SaveOfflineCatalog("off")
	}
	return m
}

// bannerVisible reports whether the first-launch offline-catalog offer should
// be shown: only on Home, only while consent is still undecided, and only
// until the user dismisses it with "not now" for this session.
func (m Model) bannerVisible() bool {
	return m.view == viewHome && m.offlineCatalog == "ask" && !m.bannerDismissed
}

// shouldOfferCatalogHint reports whether a weak search result should offer to
// enable the offline catalog: only while consent is undecided, only for a
// real query that still found nothing.
func (m Model) shouldOfferCatalogHint(query string, results int) bool {
	return m.offlineCatalog != "on" && results == 0 &&
		len([]rune(strings.TrimSpace(query))) >= minSearchLen
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case stationsMsg:
		m.loading = false
		if m.view == viewHome {
			// Home's Popular preview loads async; keep the recent section pinned on top.
			m.stations = append(append([]domain.Station{}, m.homeRecents...), msg.stations...)
		} else {
			m.stations = msg.stations
		}
		m.cursor = 0
		m.markFavorites()
		return m, nil
	case errMsg:
		m.loading = false
		m.status = "error: " + msg.err.Error()
		return m, nil
	case corpusProgressMsg:
		m.downloaded = msg.downloaded
		return m, listenProgressCmd(m.progress) // keep listening
	case corpusRefreshedMsg:
		m.refreshing = false
		m.downloaded = 0
		if msg.err != nil {
			m.status = "couldn't refresh — keeping current stations"
			return m, nil
		}
		m.status = "station list updated"
		// Re-pull the vote-sorted popular preview only where it's actually shown:
		// on Home ONLY when there are no favorites (otherwise Home shows the
		// favorites list — don't clobber it), and on the popular browse view.
		if (m.view == viewHome && len(m.favKeys) == 0) || (m.view == viewBrowse && m.crumb == "popular") {
			return m.load(popularCmd(m.dir))
		}
		return m, nil
	case titleMsg:
		m.nowTitle = sanitizeTitle(msg.title)
		return m, nil
	case playingMsg:
		// core-idle just went false → audio is actually flowing.
		m.phase = phasePlaying
		m.isPlaying = true
		return m, nil
	case idleMsg:
		// Only a real end/stop while playing returns us to idle; ignore the
		// transient idle that precedes playback during connecting.
		if m.phase == phasePlaying {
			m.phase = phaseIdle
			m.isPlaying = false
		}
		return m, nil
	case playErrMsg:
		m.phase = phaseFailed
		m.playErr = "couldn't connect — the stream may be offline"
		return m, nil
	case connectTimeoutMsg:
		if msg.attempt == m.playAttempt && m.phase == phaseConnecting {
			m.phase = phaseFailed
			m.playErr = "still connecting — the stream may be slow or offline"
		}
		return m, nil
	case updateMsg:
		m.update = msg.status
		return m, nil
	case updateAppliedMsg:
		m.updateApplying = false
		if msg.err != nil {
			m.status = "update failed: " + msg.err.Error()
		} else {
			m.status = "updated to " + m.update.Latest + " — restart onda to apply"
		}
		return m, nil
	case searchDebounceMsg:
		// Only the latest keystroke's tick, still in the search view, searches.
		// A toggle to enter-to-search mid-type can leave a tick in flight.
		if !m.liveSearch || m.view != viewSearch || msg.seq != m.searchSeq {
			return m, nil
		}
		q := strings.TrimSpace(m.search.Value())
		if len([]rune(q)) < minSearchLen {
			m.stations = nil // clear stale preview
			return m, nil
		}
		m.loading = true
		return m, tea.Batch(searchCmd(m.dir, q), m.sp.Tick)
	case spinner.TickMsg:
		if !m.loading && !m.refreshing {
			return m, nil
		}
		var cmd tea.Cmd
		m.sp, cmd = m.sp.Update(msg)
		return m, cmd
	case tea.MouseMsg:
		return m.handleMouse(msg)
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

// handleMouse maps wheel/click/hover to list actions. Input views (search,
// add, settings) keep their keyboard focus and ignore the mouse.
func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	switch m.view {
	case viewSearch, viewAdd, viewSettings:
		return m, nil
	}
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		if m.cursor > 0 {
			m.cursor--
		}
	case tea.MouseButtonWheelDown:
		if m.cursor < len(m.stations)-1 {
			m.cursor++
		}
	case tea.MouseButtonLeft:
		if msg.Action == tea.MouseActionPress {
			if idx := m.stationAtY(msg.Y); idx >= 0 {
				if idx == m.cursor {
					return m.playSelected() // click the selected row again → play
				}
				m.cursor = idx
			}
		}
	default:
		if msg.Action == tea.MouseActionMotion {
			m.hoverIdx = m.stationAtY(msg.Y)
		}
	}
	return m, nil
}

// stationAtY maps a screen row to a visible station index, or -1 if the row
// isn't over a station. Geometry mirrors viewList/viewHome layout.
func (m Model) stationAtY(y int) int {
	// Home with a pinned "recent" section on top: two sub-lists, distinct geometry.
	if recN := m.homeRecentsN(); recN > 0 {
		listRows := m.height - 14
		if listRows < 3 {
			listRows = 3
		}
		dispRecN, favStart, favEnd, _ := m.homeFavWindow(listRows)
		const recStartY = 11 // header(2)+blank(1)+panel(5)+hint(1)+blank(1)+recent-label(1)
		if y >= recStartY && y < recStartY+dispRecN {
			return y - recStartY
		}
		favRowStartY := recStartY + dispRecN + 1 // +1 for the favorites/popular label
		if j := favStart + (y - favRowStartY); y >= favRowStartY && j >= favStart && j < favEnd {
			if idx := recN + j; idx < len(m.stations) {
				return idx
			}
		}
		return -1
	}

	rowStartY := 3 // header(2) + blank(1)
	listRows := m.height - chromeHeight
	if m.view == viewHome {
		rowStartY = 11 // header(2)+blank(1)+panel(5)+hint(1)+blank(1)+label(1)
		listRows = m.height - 13
	}
	if listRows < 3 {
		listRows = 3
	}
	start, end := windowBounds(m.cursor, len(m.stations), listRows)
	idx := start + (y - rowStartY)
	if y >= rowStartY && idx >= start && idx < end && idx < len(m.stations) {
		return idx
	}
	return -1
}

func (m Model) handleKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Dedicated views capture their own keys.
	switch m.view {
	case viewSearch:
		return m.updateSearch(k)
	case viewAdd:
		return m.updateAdd(k)
	case viewSettings:
		return m.updateSettings(k)
	}

	// The first-launch offline-catalog offer intercepts y/n/esc while it's up,
	// on top of (not instead of) the normal Home key handling below.
	if m.bannerVisible() {
		switch k.String() {
		case "y":
			return m.enableCatalog()
		case "n", "esc":
			m.bannerDismissed = true
			return m, nil
		}
	}

	// Browse + favorites: list navigation.
	switch k.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "down", "j":
		if m.cursor < len(m.stations)-1 {
			m.cursor++
		}
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "enter":
		return m.playSelected()
	case "s":
		_ = m.player.Stop()
		m.isPlaying = false
		m.phase = phaseIdle
		m.playErr = ""
		m.status = "stopped"
	case "+", "=":
		return m.changeVolume(5)
	case "-", "_":
		return m.changeVolume(-5)
	case "[":
		return m.changeVariant(-1) // higher quality
	case "]":
		return m.changeVariant(1) // lower quality
	case "f":
		return m.toggleFavorite()
	case "u":
		if m.update.SelfUpdatable && !m.updateApplying {
			m.updateApplying = true
			return m, applyUpdateCmd(m.update)
		}
	case "U": // dismiss the update banner
		m.updateDismiss = true
		return m, nil
	case "/":
		m.view = viewSearch
		m.search.SetValue("")
		m.search.Focus()
		m.hoverIdx = -1
		m.stations = nil // start the live-search preview empty
		return m, nil
	case "F":
		return m.showFavorites()
	case "r":
		return m.showRecents()
	case "c":
		if m.view == viewRecents {
			return m.clearRecents()
		}
	case "p", "P":
		m.view = viewBrowse
		m.crumb = "popular"
		return m.load(popularCmd(m.dir))
	case "R":
		if !m.refreshing {
			return m.startRefresh()
		}
		return m, nil
	case "a":
		m.view = viewAdd
		m.addFocus = 0
		m.addName.SetValue("")
		m.addURL.SetValue("")
		m.addBr.SetValue("")
		m.focusAdd()
		return m, nil
	case ",":
		m.view = viewSettings
		return m, nil
	case "esc":
		return m.goHome()
	}
	return m, nil
}

// goHome returns to the Home view: opt-in recents on top, then favorites if any,
// else a Popular preview.
func (m Model) goHome() (tea.Model, tea.Cmd) {
	m.view = viewHome
	m.crumb = "home"
	m.cursor = 0
	m.homeRecents = m.recentsForHome()
	if m.store != nil {
		if favs, err := m.store.Favorites(); err == nil && len(favs) > 0 {
			m.stations = append(append([]domain.Station{}, m.homeRecents...), favs...)
			m.markFavorites()
			return m, nil
		}
	}
	// No favorites → keep recents pinned and load a Popular preview for the rest.
	m.stations = append([]domain.Station{}, m.homeRecents...)
	m.markFavorites()
	return m.load(popularCmd(m.dir))
}

// recentsForHome returns the capped, newest-first play history shown as the
// "recent" section on Home, or nil when the user hasn't opted into history.
func (m Model) recentsForHome() []domain.Station {
	if !m.history || m.store == nil {
		return nil
	}
	rec, err := m.store.Recents()
	if err != nil || len(rec) == 0 {
		return nil
	}
	if len(rec) > homeRecentsCap {
		rec = rec[:homeRecentsCap]
	}
	return rec
}

// homeRecentsN is the number of leading rows in m.stations that make up the
// Home "recent" section. Zero unless we're on Home with recorded history.
func (m Model) homeRecentsN() int {
	if m.view != viewHome {
		return 0
	}
	n := len(m.homeRecents)
	if n > len(m.stations) {
		n = len(m.stations)
	}
	return n
}

func (m Model) playSelected() (tea.Model, tea.Cmd) {
	if len(m.stations) == 0 {
		return m, nil
	}
	st := m.stations[m.cursor]

	// Re-playing the station that's already playing keeps the bitrate you picked
	// with the [ ] chooser; switching to a different station uses your preference.
	if m.isPlaying && favKey(st) == favKey(m.playing) && m.varIdx >= 0 && m.varIdx < len(m.playing.Variants) {
		v := m.playing.Variants[m.varIdx]
		_ = m.player.Play(v.URL)
		m.nowTitle = ""
		m.status = "playing " + m.playing.Name + " · " + v.Quality()
		m.recordRecent(st)
		return m.startConnecting()
	}

	if v, ok := st.SelectVariant(m.quality); ok {
		m.playing = st
		m.varIdx = indexOfVariant(st.Variants, v)
		_ = m.player.Play(v.URL)
		m.isPlaying = true
		m.nowTitle = ""
		m.status = "playing " + st.Name + " · " + v.Quality()
		m.recordRecent(st)
		return m.startConnecting()
	}
	m.status = "no playable stream for " + st.Name
	return m, nil
}

// recordRecent appends the station to the local play history, but only when the
// user has opted in (history setting). Stored locally only — see store.AddRecent.
func (m Model) recordRecent(st domain.Station) {
	if m.history && m.store != nil {
		_ = m.store.AddRecent(st)
	}
}

// startConnecting enters the connecting phase and schedules a stale-guarded
// timeout so a stream that never starts is reported as failed, not stuck.
func (m Model) startConnecting() (tea.Model, tea.Cmd) {
	m.phase = phaseConnecting
	m.playErr = ""
	m.playAttempt++
	attempt := m.playAttempt
	return m, tea.Tick(connectTimeout, func(time.Time) tea.Msg {
		return connectTimeoutMsg{attempt: attempt}
	})
}

// changeVariant switches the playing station to another available bitrate.
// delta -1 selects higher quality (variants are sorted best-first), +1 lower.
func (m Model) changeVariant(delta int) (tea.Model, tea.Cmd) {
	if !m.isPlaying || len(m.playing.Variants) < 2 {
		m.status = "only one quality available"
		return m, nil
	}
	m.varIdx += delta
	if m.varIdx < 0 {
		m.varIdx = 0
	}
	if m.varIdx > len(m.playing.Variants)-1 {
		m.varIdx = len(m.playing.Variants) - 1
	}
	v := m.playing.Variants[m.varIdx]
	_ = m.player.Play(v.URL)
	m.status = "quality " + v.Quality()
	return m.startConnecting()
}

func indexOfVariant(vs []domain.StreamVariant, target domain.StreamVariant) int {
	for i, v := range vs {
		if v.URL == target.URL {
			return i
		}
	}
	return 0
}

func (m Model) changeVolume(delta int) (tea.Model, tea.Cmd) {
	m.volume = clampVolume(m.volume + delta)
	_ = m.player.Volume(m.volume)
	if m.store != nil {
		_ = m.store.SaveVolume(m.volume)
	}
	m.status = "volume " + strconv.Itoa(m.volume) + "%"
	return m, nil
}

// clampVolume bounds a volume to the 0–100 range.
func clampVolume(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}

func (m Model) toggleFavorite() (Model, tea.Cmd) {
	if len(m.stations) == 0 {
		return m, nil
	}
	st := m.stations[m.cursor]
	fav, _ := m.store.IsFavorite(st)
	if fav {
		_ = m.store.RemoveFavorite(st)
		m.status = "removed from favorites: " + st.Name
	} else {
		_ = m.store.AddFavorite(st)
		m.status = "added to favorites: " + st.Name
	}
	m.markFavorites()
	return m, nil
}

func (m Model) showFavorites() (tea.Model, tea.Cmd) {
	favs, err := m.store.Favorites()
	if err != nil {
		m.status = "error loading favorites: " + err.Error()
		return m, nil
	}
	m.view = viewFavorites
	m.stations = favs
	m.cursor = 0
	m.markFavorites()
	return m, nil
}

// showRecents opens the locally-stored play history (newest first).
func (m Model) showRecents() (tea.Model, tea.Cmd) {
	if m.store == nil {
		return m, nil
	}
	recents, err := m.store.Recents()
	if err != nil {
		m.status = "error loading recents: " + err.Error()
		return m, nil
	}
	m.view = viewRecents
	m.crumb = "recents"
	m.stations = recents
	m.cursor = 0
	m.markFavorites()
	return m, nil
}

// clearRecents wipes the local play history.
func (m Model) clearRecents() (tea.Model, tea.Cmd) {
	if m.store != nil {
		_ = m.store.ClearRecents()
	}
	m.stations = nil
	m.cursor = 0
	m.status = "cleared play history"
	return m, nil
}

// markFavorites refreshes the set of favorited station keys for ★ rendering.
func (m *Model) markFavorites() {
	m.favKeys = map[string]bool{}
	if m.store == nil {
		return
	}
	favs, err := m.store.Favorites()
	if err != nil {
		return
	}
	for _, f := range favs {
		m.favKeys[favKey(f)] = true
	}
}

func favKey(s domain.Station) string { return s.Name + "|" + s.Homepage }

func (m Model) cycleQuality() Model {
	switch m.quality {
	case domain.QualityHighest:
		m.quality = domain.QualityBalanced
	case domain.QualityBalanced:
		m.quality = domain.QualityLowest
	default:
		m.quality = domain.QualityHighest
	}
	return m
}

func (m Model) cycleTracking() Model {
	switch m.tracking {
	case "never":
		m.tracking = "opt-in"
	case "opt-in":
		m.tracking = "opt-out"
	default:
		m.tracking = "never"
	}
	return m
}

func (m Model) cycleTheme() Model {
	t := nextTheme(m.themeName)
	m.themeName = t.Name
	m.st = newStyles(t)
	return m
}

func (m *Model) focusAdd() {
	m.addName.Blur()
	m.addURL.Blur()
	m.addBr.Blur()
	switch m.addFocus {
	case 0:
		m.addName.Focus()
	case 1:
		m.addURL.Focus()
	case 2:
		m.addBr.Focus()
	}
}

func (m *Model) blurAdd() {
	m.addName.Blur()
	m.addURL.Blur()
	m.addBr.Blur()
}

func (m Model) submitAdd() (tea.Model, tea.Cmd) {
	name := strings.TrimSpace(m.addName.Value())
	url := strings.TrimSpace(m.addURL.Value())
	if name == "" || url == "" {
		m.status = "name and URL are required"
		return m, nil
	}
	br, _ := strconv.Atoi(strings.TrimSpace(m.addBr.Value()))
	st := domain.Station{
		Name:     name,
		Country:  "Custom",
		Variants: []domain.StreamVariant{{URL: url, Bitrate: br}},
	}
	if err := m.store.AddCustom(st); err != nil {
		m.status = "error saving: " + err.Error()
		return m, nil
	}
	m.view = viewBrowse
	m.blurAdd()
	m.status = "added " + name
	return m, nil
}

func (m Model) View() string {
	var s string
	switch m.view {
	case viewHome:
		s = m.viewHome()
	case viewSearch:
		s = m.viewSearch()
	case viewAdd:
		s = m.viewAdd()
	case viewSettings:
		s = m.viewSettings()
	default:
		s = m.viewList()
	}
	return indentLines(s, gutter)
}
