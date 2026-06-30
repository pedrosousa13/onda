package tui

import (
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	searchDebounce = 350 * time.Millisecond
	minSearchLen   = 2
)

// searchDebounceCmd fires a debounce tick tagged with seq; only the latest seq
// (i.e. no newer keystroke) actually triggers a search.
func searchDebounceCmd(seq int) tea.Cmd {
	return tea.Tick(searchDebounce, func(time.Time) tea.Msg {
		return searchDebounceMsg{seq: seq}
	})
}

func (m Model) updateSearch(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "esc":
		m.search.Blur()
		return m.goHome()
	case "enter":
		q := m.search.Value()
		m.view = viewBrowse
		m.search.Blur()
		m.crumb = "“" + q + "”"
		return m.load(searchCmd(m.dir, q))
	}
	var cmd tea.Cmd
	m.search, cmd = m.search.Update(k)
	// Live search: schedule a debounced search tagged with this keystroke's seq.
	m.searchSeq++
	return m, tea.Batch(cmd, searchDebounceCmd(m.searchSeq))
}

func (m Model) viewSearch() string {
	var b strings.Builder
	b.WriteString(m.header("search"))
	b.WriteString("\n\n")
	b.WriteString("  " + m.search.View() + "\n\n")

	q := strings.TrimSpace(m.search.Value())
	switch {
	case len([]rune(q)) < minSearchLen:
		b.WriteString(m.st.Help.Render("  keep typing — matches name, country, or genre") + "\n")
	case m.loading:
		b.WriteString(m.st.Meta.Render("  "+m.sp.View()+" searching…") + "\n")
	case len(m.stations) == 0:
		b.WriteString(m.st.Meta.Render("  no matches for “"+q+"”") + "\n")
	default:
		const preview = 8
		for i, s := range m.stations {
			if i >= preview {
				b.WriteString(m.st.Help.Render("  …and "+strconv.Itoa(len(m.stations)-preview)+" more — press ") +
					m.st.Key.Render("⏎") + m.st.Help.Render(" to open") + "\n")
				break
			}
			b.WriteString(m.renderRow(m.contentWidth(), i, s) + "\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(m.st.Help.Render("  ") + m.st.Key.Render("⏎") + m.st.Help.Render(" open results  ·  ") +
		m.st.Key.Render("esc") + m.st.Help.Render(" cancel") + "\n")
	return b.String()
}
