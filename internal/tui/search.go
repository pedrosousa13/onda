package tui

import tea "github.com/charmbracelet/bubbletea"

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
		m.status = "searching: " + q
		return m, searchCmd(m.dir, q)
	}
	var cmd tea.Cmd
	m.search, cmd = m.search.Update(k)
	return m, cmd
}

func (m Model) viewSearch() string {
	return titleStyle.Render("search") + "\n\n" +
		m.search.View() + "\n\n" +
		statusStyle.Render("enter: search · esc: cancel")
}
