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
	case "up":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case "down":
		if m.cursor < len(m.stations)-1 {
			m.cursor++
		}
		return m, nil
	case "enter":
		// Play the highlighted live result and drop into the browse list so the
		// now-playing panel shows. If results haven't loaded yet, search first.
		q := strings.TrimSpace(m.search.Value())
		m.search.Blur()
		m.view = viewBrowse
		m.crumb = "“" + q + "”"
		// In live mode a result is already highlighted, so enter plays it. In
		// enter-to-search mode the list holds stale stations, so always search.
		if m.liveSearch && len(m.stations) > 0 {
			return m.playSelected()
		}
		return m.load(searchCmd(m.dir, q))
	case "ctrl+o":
		if m.shouldOfferCatalogHint(strings.TrimSpace(m.search.Value()), len(m.stations)) {
			return m.enableCatalog()
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.search, cmd = m.search.Update(k)
	if !m.liveSearch {
		return m, cmd // enter-to-search: don't query as the user types
	}
	// Live search: schedule a debounced search tagged with this keystroke's seq.
	m.searchSeq++
	return m, tea.Batch(cmd, searchDebounceCmd(m.searchSeq))
}

func (m Model) viewSearch() string {
	var b strings.Builder
	b.WriteString(m.header("search"))
	b.WriteString("\n\n")
	b.WriteString("  " + m.search.View() + "\n\n")

	if !m.liveSearch {
		b.WriteString(m.st.Help.Render("  ") + m.st.Key.Render("⏎") +
			m.st.Help.Render(" search  ·  ") + m.st.Key.Render("esc") +
			m.st.Help.Render(" cancel") + "\n")
		return b.String()
	}

	q := strings.TrimSpace(m.search.Value())
	footer := m.st.Help.Render("  ") + m.st.Key.Render("esc") + m.st.Help.Render(" cancel")
	switch {
	case len([]rune(q)) < minSearchLen:
		b.WriteString(m.st.Help.Render("  keep typing — matches name, country, or genre") + "\n")
	case m.loading:
		b.WriteString(m.st.Meta.Render("  "+m.sp.View()+" searching…") + "\n")
	case len(m.stations) == 0:
		b.WriteString(m.st.Meta.Render("  no matches for “"+q+"”") + "\n")
		if m.shouldOfferCatalogHint(q, len(m.stations)) {
			b.WriteString(m.st.Meta.Render("  ⓘ enable the full catalog to catch typos — "+catalogSizeHint+"  ") +
				m.st.Key.Render("[ctrl+o]") + "\n")
		}
	default:
		const preview = 8
		start, end := windowBounds(m.cursor, len(m.stations), preview)
		for i := start; i < end; i++ {
			b.WriteString(m.renderRow(m.contentWidth(), i, m.stations[i]) + "\n")
		}
		footer = m.st.Help.Render("  "+strconv.Itoa(len(m.stations))+" matches  ·  ") +
			m.st.Key.Render("↑↓") + m.st.Help.Render(" pick  ·  ") +
			m.st.Key.Render("⏎") + m.st.Help.Render(" play  ·  ") +
			m.st.Key.Render("esc") + m.st.Help.Render(" cancel")
	}

	b.WriteString("\n")
	b.WriteString(footer + "\n")
	return b.String()
}
