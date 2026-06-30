package tui

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pedrosousa13/radio/internal/domain"
)

// reserved vertical space: header(2) + blank(1) + now-panel(5) + footer(1).
const chromeHeight = 9

func (m Model) viewList() string {
	crumb := m.crumb
	if crumb == "" {
		crumb = "all stations"
	}
	if m.view == viewFavorites {
		crumb = "favorites"
	}
	if m.loading {
		crumb = m.sp.View() + " loading"
	}

	var b strings.Builder
	b.WriteString(m.header(crumb))
	b.WriteString("\n\n")

	if m.loading && len(m.stations) == 0 {
		b.WriteString(m.st.Meta.Render("  "+m.sp.View()+" finding stations…") + "\n")
		for i := 1; i < m.height-chromeHeight; i++ {
			b.WriteString("\n")
		}
		b.WriteString(m.nowPanel())
		b.WriteString("\n")
		b.WriteString(m.footer())
		return b.String()
	}

	listRows := m.height - chromeHeight
	if listRows < 3 {
		listRows = 3
	}

	if len(m.stations) == 0 {
		b.WriteString(m.st.Meta.Render("  Nothing here yet.") + "\n")
		hint := m.st.Help.Render("  press ") + m.st.Key.Render("/") +
			m.st.Help.Render(" to search the world, or ") + m.st.Key.Render("a") +
			m.st.Help.Render(" to add your own stream")
		b.WriteString(hint + "\n")
		for i := 2; i < listRows; i++ {
			b.WriteString("\n")
		}
	} else {
		start, end := windowBounds(m.cursor, len(m.stations), listRows)
		for i := start; i < end; i++ {
			b.WriteString(m.renderRow(m.width, i, m.stations[i]) + "\n")
		}
		for i := end - start; i < listRows; i++ {
			b.WriteString("\n")
		}
	}

	b.WriteString(m.nowPanel())
	b.WriteString("\n")
	b.WriteString(m.footer())
	return b.String()
}

// header renders the title line plus a right-aligned view label and a rule.
func (m Model) header(crumb string) string {
	left := m.st.Title.Render("radio") + m.st.Subtitle.Render("  ·  wander the world")
	right := m.st.Crumb.Render(crumb)
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	rule := m.st.Rule.Render(strings.Repeat("─", max(1, m.width)))
	return left + strings.Repeat(" ", gap) + right + "\n" + rule
}

// renderRow lays out one station: ▌ name … country · 128k ★
func (m Model) renderRow(w, idx int, s domain.Station) string {
	sel := idx == m.cursor

	meta := s.Country
	if v, ok := s.SelectVariant(m.quality); ok {
		if q := v.Quality(); q != "—" {
			meta += " · " + q
		}
	}
	fav := m.favKeys[favKey(s)]
	starPlain := ""
	if fav {
		starPlain = " ★"
	}
	rightPlain := meta + starPlain

	avail := w - 2 /*marker*/ - 1 /*gap*/ - lipgloss.Width(rightPlain)
	if avail < 4 {
		avail = 4
	}
	name := truncate(s.Name, avail)
	pad := avail - lipgloss.Width(name)
	if pad < 0 {
		pad = 0
	}

	var marker, nameS string
	if sel {
		marker = m.st.SelBar.Render("▌ ")
		nameS = m.st.SelName.Render(name)
	} else {
		marker = "  "
		nameS = m.st.Item.Render(name)
	}
	starS := ""
	if fav {
		starS = " " + m.st.Star.Render("★")
	}
	return marker + nameS + strings.Repeat(" ", pad) + " " + m.st.Meta.Render(meta) + starS
}

// nowPanel is the bordered "now playing" hero pinned below the list.
func (m Model) nowPanel() string {
	inner := m.width - 4 // border(2) + padding(2)
	if inner < 12 {
		inner = 12
	}

	var title, sub, tags string
	if m.isPlaying {
		title = m.st.NowTitle.Render("♫ " + truncate(m.playing.Name, inner-2))
		if m.nowTitle != "" {
			sub = m.st.NowText.Render(truncate(m.nowTitle, inner))
		} else {
			sub = m.st.Meta.Render("live")
		}
		tags = strings.Join(m.playing.Tags, ", ")
	} else {
		title = m.st.NowTitle.Render("♫ nothing playing")
		sub = m.st.Meta.Render("select a station and press enter")
	}

	vol := m.volumeBar()
	// Left of the volume bar: the bitrate chooser when there's a choice, else tags.
	left := m.st.Meta.Render(truncate(tags, max(0, inner-lipgloss.Width(vol)-1)))
	if m.isPlaying && len(m.playing.Variants) > 1 {
		left = m.qualityChips(inner - lipgloss.Width(vol) - 1)
	}
	gap := inner - lipgloss.Width(left) - lipgloss.Width(vol)
	if gap < 1 {
		gap = 1
	}
	third := left + strings.Repeat(" ", gap) + vol

	content := title + "\n" + sub + "\n" + third
	// Panel.Width is the content box incl. padding(0,1); total = width-2+border(2) = width.
	return m.st.Panel.Width(m.width - 2).Render(content)
}

// qualityChips renders the playing station's available bitrates, the active one
// highlighted, prefixed with the [ ] hint. Truncates to maxW columns.
func (m Model) qualityChips(maxW int) string {
	out := m.st.Help.Render("[ ] ")
	for i, v := range m.playing.Variants {
		chip := v.Quality()
		if i == m.varIdx {
			chip = m.st.Crumb.Render(chip)
		} else {
			chip = m.st.Meta.Render(chip)
		}
		next := out + chip + " "
		if lipgloss.Width(next) > maxW {
			break
		}
		out = next
	}
	return strings.TrimRight(out, " ")
}

func (m Model) volumeBar() string {
	const cells = 10
	on := m.volume * cells / 100
	if on > cells {
		on = cells
	}
	bar := m.st.VolOn.Render(strings.Repeat("▮", on)) +
		m.st.VolOff.Render(strings.Repeat("▯", cells-on))
	return bar + " " + m.st.Meta.Render(strconv.Itoa(m.volume)+"%")
}

func (m Model) footer() string {
	pairs := [][2]string{
		{"↑↓", "move"}, {"⏎", "play"}, {"s", "stop"}, {"+/-", "vol"},
		{"[ ]", "quality"}, {"f", "★"}, {"F", "favs"}, {"/", "search"},
		{"a", "add"}, {",", "settings"}, {"q", "quit"},
	}
	sep := m.st.Help.Render("  ")
	out := ""
	wsum := 0
	for i, p := range pairs {
		seg := m.st.Key.Render(p[0]) + " " + m.st.Help.Render(p[1])
		add := lipgloss.Width(seg)
		if i > 0 {
			add += 2
		}
		if wsum+add > m.width {
			break
		}
		if i > 0 {
			out += sep
		}
		out += seg
		wsum += add
	}
	return out
}

// windowBounds returns the visible [start,end) slice that keeps cursor in view.
func windowBounds(cursor, n, rows int) (int, int) {
	if n <= rows {
		return 0, n
	}
	start := cursor - rows/2
	if start < 0 {
		start = 0
	}
	if start+rows > n {
		start = n - rows
	}
	return start, start + rows
}

// truncate shortens s to at most w display columns, adding an ellipsis.
func truncate(s string, w int) string {
	if w <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= w {
		return s
	}
	if w == 1 {
		return "…"
	}
	return string(r[:w-1]) + "…"
}
