package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) updateSettings(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "esc", "q", ",":
		return m.goHome()
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
	case "4":
		m = m.cycleTheme()
		if m.store != nil {
			_ = m.store.SaveTheme(m.themeName)
		}
	case "5":
		m.updateCheck = !m.updateCheck
		if m.store != nil {
			_ = m.store.SaveUpdateCheck(m.updateCheck)
		}
	case "6":
		m.liveSearch = !m.liveSearch
		if m.store != nil {
			_ = m.store.SaveLiveSearch(m.liveSearch)
		}
	case "7":
		m.normalize = !m.normalize
		if m.player != nil {
			_ = m.player.SetNormalize(m.normalize)
		}
		if m.store != nil {
			_ = m.store.SaveNormalize(m.normalize)
		}
	case "8":
		if m.offlineCatalog == "on" {
			m = m.disableCatalog()
		} else {
			return m.enableCatalog()
		}
	}
	return m, nil
}

func (m Model) viewSettings() string {
	var b strings.Builder
	b.WriteString(m.header("settings"))
	b.WriteString("\n\n")

	row := func(key, label, value string) string {
		pad := 22 - lipgloss.Width(label)
		if pad < 1 {
			pad = 1
		}
		return "  " + m.st.Key.Render(key) + "  " + m.st.Item.Render(label) +
			strings.Repeat(" ", pad) + m.st.Crumb.Render(value)
	}

	b.WriteString(row("1", "audio quality", string(m.quality)) + "\n")
	b.WriteString(row("2", "popularity tracking", m.tracking) + "\n")
	b.WriteString(row("3", "play history", fmt.Sprintf("%v", m.history)) + "\n")
	b.WriteString(row("4", "theme", m.themeName) + "\n")
	b.WriteString(row("5", "check for updates", fmt.Sprintf("%v", m.updateCheck)) + "\n")
	b.WriteString(row("6", "live search", fmt.Sprintf("%v", m.liveSearch)) + "\n")
	b.WriteString(row("7", "loudness normalization", fmt.Sprintf("%v", m.normalize)) + "\n")

	catalogState := "off"
	if m.offlineCatalog == "on" {
		catalogState = "on"
	}
	b.WriteString(row("8", "offline catalog", fmt.Sprintf("%s (%s)", catalogState, catalogSizeHint)) + "\n\n")

	b.WriteString(m.st.Help.Render("  press a number to change · ") +
		m.st.Key.Render("esc") + m.st.Help.Render(" back") + "\n")
	b.WriteString(m.st.Help.Render("  tracking ‘never’ (default) reports nothing about what you play") + "\n")
	b.WriteString(m.st.Help.Render("  live search off → queries sent only when you press ⏎") + "\n")
	return b.String()
}
