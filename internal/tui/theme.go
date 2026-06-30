package tui

import "github.com/charmbracelet/lipgloss"

// Theme is a named palette. Colors are hex strings rendered via lipgloss,
// which adapts them to the terminal's color profile.
type Theme struct {
	Name   string
	Bg     string // panel/base background
	Fg     string // primary text
	Dim    string // subtext, help, secondary
	Sel    string // selection accent (bar + selected text)
	SelBg  string // selected row background
	Accent string // headers, now-playing title
	Good   string // favorite star, "playing"
	Border string // panel borders
}

// themes is the ordered registry; the first is the default.
var themes = []Theme{
	{
		Name: "catppuccin-mocha",
		Bg:   "#1e1e2e", Fg: "#cdd6f4", Dim: "#a6adc8",
		Sel: "#cba6f7", SelBg: "#313244", Accent: "#89b4fa",
		Good: "#a6e3a1", Border: "#585b70",
	},
	{
		Name: "catppuccin-macchiato",
		Bg:   "#24273a", Fg: "#cad3f5", Dim: "#a5adcb",
		Sel: "#c6a0f6", SelBg: "#363a4f", Accent: "#8aadf4",
		Good: "#a6da95", Border: "#5b6078",
	},
	{
		Name: "catppuccin-frappe",
		Bg:   "#303446", Fg: "#c6d0f5", Dim: "#a5adce",
		Sel: "#ca9ee6", SelBg: "#414559", Accent: "#8caaee",
		Good: "#a6d189", Border: "#626880",
	},
	{
		Name: "catppuccin-latte",
		Bg:   "#eff1f5", Fg: "#4c4f69", Dim: "#6c6f85",
		Sel: "#8839ef", SelBg: "#ccd0da", Accent: "#1e66f5",
		Good: "#40a02b", Border: "#acb0be",
	},
	{
		Name: "dracula",
		Bg:   "#282a36", Fg: "#f8f8f2", Dim: "#6272a4",
		Sel: "#bd93f9", SelBg: "#44475a", Accent: "#8be9fd",
		Good: "#50fa7b", Border: "#6272a4",
	},
	{
		Name: "nord",
		Bg:   "#2e3440", Fg: "#d8dee9", Dim: "#616e88",
		Sel: "#88c0d0", SelBg: "#3b4252", Accent: "#81a1c1",
		Good: "#a3be8c", Border: "#4c566a",
	},
	{
		Name: "gruvbox",
		Bg:   "#282828", Fg: "#ebdbb2", Dim: "#a89984",
		Sel: "#fabd2f", SelBg: "#3c3836", Accent: "#83a598",
		Good: "#b8bb26", Border: "#665c54",
	},
}

// themeByName returns the named theme, or the default (first) if not found.
func themeByName(name string) Theme {
	for _, t := range themes {
		if t.Name == name {
			return t
		}
	}
	return themes[0]
}

// nextTheme returns the theme after the named one (wrapping), for cycling.
func nextTheme(name string) Theme {
	for i, t := range themes {
		if t.Name == name {
			return themes[(i+1)%len(themes)]
		}
	}
	return themes[0]
}

func (t Theme) color(hex string) lipgloss.Color { return lipgloss.Color(hex) }
