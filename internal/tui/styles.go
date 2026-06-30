package tui

import "github.com/charmbracelet/lipgloss"

// Styles holds every lipgloss style used by the views, derived from a Theme.
type Styles struct {
	Title    lipgloss.Style
	Subtitle lipgloss.Style
	Crumb    lipgloss.Style // current-view label in the header
	Item     lipgloss.Style
	SelName  lipgloss.Style
	SelBar   lipgloss.Style
	Meta     lipgloss.Style
	Star     lipgloss.Style
	Panel    lipgloss.Style
	NowTitle lipgloss.Style
	NowText  lipgloss.Style
	VolOn    lipgloss.Style
	VolOff   lipgloss.Style
	Help     lipgloss.Style
	Key      lipgloss.Style
	Input    lipgloss.Style
	Rule     lipgloss.Style
}

func newStyles(t Theme) Styles {
	c := func(hex string) lipgloss.Color { return lipgloss.Color(hex) }
	return Styles{
		Title:    lipgloss.NewStyle().Bold(true).Foreground(c(t.Accent)),
		Subtitle: lipgloss.NewStyle().Foreground(c(t.Dim)),
		Crumb:    lipgloss.NewStyle().Bold(true).Foreground(c(t.Sel)),
		Item:     lipgloss.NewStyle().Foreground(c(t.Fg)),
		SelName:  lipgloss.NewStyle().Bold(true).Foreground(c(t.Sel)),
		SelBar:   lipgloss.NewStyle().Foreground(c(t.Sel)),
		Meta:     lipgloss.NewStyle().Foreground(c(t.Dim)),
		Star:     lipgloss.NewStyle().Foreground(c(t.Good)),
		Panel: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(c(t.Border)).
			Padding(0, 1),
		NowTitle: lipgloss.NewStyle().Bold(true).Foreground(c(t.Accent)),
		NowText:  lipgloss.NewStyle().Foreground(c(t.Fg)),
		VolOn:    lipgloss.NewStyle().Foreground(c(t.Good)),
		VolOff:   lipgloss.NewStyle().Foreground(c(t.Border)),
		Help:     lipgloss.NewStyle().Foreground(c(t.Dim)),
		Key:      lipgloss.NewStyle().Bold(true).Foreground(c(t.Accent)),
		Input:    lipgloss.NewStyle().Foreground(c(t.Fg)),
		Rule:     lipgloss.NewStyle().Foreground(c(t.Border)),
	}
}
