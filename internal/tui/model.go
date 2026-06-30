package tui

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/pedrosousa13/radio/internal/domain"
)

type view int

const (
	viewBrowse view = iota
	viewSearch
	viewFavorites
	viewAdd
	viewSettings
)

type Player interface {
	Play(url string) error
	Stop() error
	Volume(pct int) error
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
}

type Model struct {
	dir       Searcher
	player    Player
	store     Store
	view      view
	stations  []domain.Station
	cursor    int
	status    string
	nowTitle  string
	playing   domain.Station
	isPlaying bool
	quality   domain.QualityPref
	tracking  string
	history   bool
	volume    int
	themeName string
	st        Styles
	width     int
	height    int
	favKeys   map[string]bool

	search   textinput.Model
	addName  textinput.Model
	addURL   textinput.Model
	addBr    textinput.Model
	addFocus int // 0=name, 1=url, 2=bitrate
}

func New(dir Searcher, p Player, st Store, quality domain.QualityPref, tracking string, history bool, theme string) Model {
	search := textinput.New()
	search.Placeholder = "search stations, country, or genre…"
	name := textinput.New()
	name.Placeholder = "name"
	url := textinput.New()
	url.Placeholder = "https://stream-url"
	br := textinput.New()
	br.Placeholder = "bitrate kbps (optional)"

	t := themeByName(theme)
	return Model{
		dir: dir, player: p, store: st,
		quality: quality, tracking: tracking, history: history,
		volume: 100, themeName: t.Name, st: newStyles(t),
		width: 80, height: 24, favKeys: map[string]bool{},
		search: search, addName: name, addURL: url, addBr: br,
	}
}

func (m Model) Init() tea.Cmd { return searchCmd(m.dir, "") }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case stationsMsg:
		m.stations = msg.stations
		m.cursor = 0
		m.markFavorites()
		return m, nil
	case errMsg:
		m.status = "error: " + msg.err.Error()
		return m, nil
	case titleMsg:
		m.nowTitle = msg.title
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
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
		m.status = "stopped"
	case "+", "=":
		return m.changeVolume(5)
	case "-", "_":
		return m.changeVolume(-5)
	case "f":
		return m.toggleFavorite()
	case "/":
		m.view = viewSearch
		m.search.SetValue("")
		m.search.Focus()
		return m, nil
	case "F":
		return m.showFavorites()
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
		m.view = viewBrowse
		return m, searchCmd(m.dir, "")
	}
	return m, nil
}

func (m Model) playSelected() (tea.Model, tea.Cmd) {
	if len(m.stations) == 0 {
		return m, nil
	}
	st := m.stations[m.cursor]
	if v, ok := st.SelectVariant(m.quality); ok {
		_ = m.player.Play(v.URL)
		m.playing = st
		m.isPlaying = true
		m.nowTitle = ""
		m.status = "playing " + st.Name
	} else {
		m.status = "no playable stream for " + st.Name
	}
	return m, nil
}

func (m Model) changeVolume(delta int) (tea.Model, tea.Cmd) {
	m.volume += delta
	if m.volume < 0 {
		m.volume = 0
	}
	if m.volume > 100 {
		m.volume = 100
	}
	_ = m.player.Volume(m.volume)
	m.status = "volume " + strconv.Itoa(m.volume) + "%"
	return m, nil
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
	switch m.view {
	case viewSearch:
		return m.viewSearch()
	case viewAdd:
		return m.viewAdd()
	case viewSettings:
		return m.viewSettings()
	default:
		return m.viewList()
	}
}
