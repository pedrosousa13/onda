package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateSearch(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "esc":
		m.view = viewBrowse
		m.search.Blur()
		return m, searchCmd(m.dir, "")
	case "enter":
		q := m.search.Value()
		m.view = viewBrowse
		m.search.Blur()
		m.status = "searching “" + q + "”"
		return m, searchCmd(m.dir, q)
	}
	var cmd tea.Cmd
	m.search, cmd = m.search.Update(k)
	return m, cmd
}

func (m Model) viewSearch() string {
	var b strings.Builder
	b.WriteString(m.header("search"))
	b.WriteString("\n\n")
	b.WriteString("  " + m.search.View() + "\n\n")
	b.WriteString(m.st.Help.Render("  ") + m.st.Key.Render("⏎") + m.st.Help.Render(" search  ·  ") +
		m.st.Key.Render("esc") + m.st.Help.Render(" cancel") + "\n")
	return b.String()
}
