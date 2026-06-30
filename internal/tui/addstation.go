package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) updateAdd(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "esc":
		m.view = viewBrowse
		m.blurAdd()
		return m, nil
	case "tab", "down":
		m.addFocus = (m.addFocus + 1) % 3
		m.focusAdd()
		return m, nil
	case "shift+tab", "up":
		m.addFocus = (m.addFocus + 2) % 3
		m.focusAdd()
		return m, nil
	case "enter":
		return m.submitAdd()
	}
	var cmd tea.Cmd
	switch m.addFocus {
	case 0:
		m.addName, cmd = m.addName.Update(k)
	case 1:
		m.addURL, cmd = m.addURL.Update(k)
	case 2:
		m.addBr, cmd = m.addBr.Update(k)
	}
	return m, cmd
}

func (m Model) viewAdd() string {
	b := titleStyle.Render("add a station") + "\n\n"
	b += "name:    " + m.addName.View() + "\n"
	b += "url:     " + m.addURL.View() + "\n"
	b += "bitrate: " + m.addBr.View() + "\n\n"
	b += statusStyle.Render(m.status) + "\n"
	b += statusStyle.Render("tab: next field · enter: save · esc: cancel") + "\n"
	return b
}
