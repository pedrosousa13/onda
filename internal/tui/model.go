package tui

import (
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
}

type Model struct {
	dir      Searcher
	player   Player
	view     view
	stations []domain.Station
	cursor   int
	status   string
	nowTitle string
	quality  domain.QualityPref
}

func New(dir Searcher, p Player, quality domain.QualityPref) Model {
	return Model{dir: dir, player: p, quality: quality}
}

func (m Model) Init() tea.Cmd { return searchCmd(m.dir, "") }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case stationsMsg:
		m.stations = msg.stations
		m.cursor = 0
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
		if len(m.stations) == 0 {
			return m, nil
		}
		st := m.stations[m.cursor]
		if v, ok := st.SelectVariant(m.quality); ok {
			_ = m.player.Play(v.URL)
			m.status = "playing: " + st.Name
		}
	case "s":
		_ = m.player.Stop()
		m.status = "stopped"
	}
	return m, nil
}

func (m Model) View() string {
	var b string
	b += titleStyle.Render("radio — wander the world") + "\n\n"
	for i, s := range m.stations {
		line := s.Name + "  " + s.Country
		if i == m.cursor {
			line = selectedStyle.Render("> " + line)
		} else {
			line = "  " + line
		}
		b += line + "\n"
	}
	b += "\n"
	if m.nowTitle != "" {
		b += "♪ " + m.nowTitle + "\n"
	}
	b += statusStyle.Render(m.status) + "\n"
	b += statusStyle.Render("↑/↓ move · enter play · s stop · q quit") + "\n"
	return b
}
