package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

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
	var b strings.Builder
	b.WriteString(m.header("add a station"))
	b.WriteString("\n\n")

	field := func(label string, ti string) string {
		return "  " + m.st.Meta.Render(label) + "  " + ti
	}
	b.WriteString(field("name   ", m.addName.View()) + "\n")
	b.WriteString(field("url    ", m.addURL.View()) + "\n")
	b.WriteString(field("bitrate", m.addBr.View()) + "\n\n")

	if m.status != "" {
		b.WriteString(m.st.Subtitle.Render("  "+m.status) + "\n")
	}
	b.WriteString(m.st.Help.Render("  ") + m.st.Key.Render("tab") + m.st.Help.Render(" next  ·  ") +
		m.st.Key.Render("⏎") + m.st.Help.Render(" save  ·  ") +
		m.st.Key.Render("esc") + m.st.Help.Render(" cancel") + "\n")
	return b.String()
}
