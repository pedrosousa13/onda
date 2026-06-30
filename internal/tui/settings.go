package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) updateSettings(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "esc", "q", ",":
		m.view = viewBrowse
		return m, nil
	case "1":
		m = m.cycleQuality()
		if m.store != nil {
			_ = m.store.SaveQuality(m.quality)
		}
	case "2":
		m = m.cycleTracking()
		if m.store != nil {
			_ = m.store.SaveTracking(m.tracking)
		}
	case "3":
		m.history = !m.history
		if m.store != nil {
			_ = m.store.SaveHistory(m.history)
		}
	}
	return m, nil
}

func (m Model) viewSettings() string {
	b := titleStyle.Render("settings") + "\n\n"
	b += fmt.Sprintf("1) quality:           %s\n", m.quality)
	b += fmt.Sprintf("2) popularity tracking: %s\n", m.tracking)
	b += fmt.Sprintf("3) play history:      %v\n\n", m.history)
	b += statusStyle.Render("press 1/2/3 to change · esc: back") + "\n"
	b += statusStyle.Render("tracking 'never' (default) reports nothing about what you play") + "\n"
	return b
}
