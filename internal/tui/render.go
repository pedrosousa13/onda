package tui

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pedrosousa13/onda/internal/domain"
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
	banner := m.updateBanner()
	reserved := chromeHeight
	if banner != "" {
		b.WriteString("\n" + banner)
		reserved++ // banner consumes one line below the header
	}
	b.WriteString("\n\n")

	if m.loading && len(m.stations) == 0 {
		b.WriteString(m.st.Meta.Render("  "+m.sp.View()+" finding stations…") + "\n")
		for i := 1; i < m.height-reserved; i++ {
			b.WriteString("\n")
		}
		b.WriteString(m.nowPanel())
		b.WriteString("\n")
		b.WriteString(m.footer())
		return b.String()
	}

	listRows := m.height - reserved
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

// viewHome is the landing screen: now-playing hero on top, then favorites
// (or a Popular preview when there are none).
func (m Model) viewHome() string {
	var b strings.Builder
	b.WriteString(m.header("home"))
	b.WriteString("\n\n")
	b.WriteString(m.nowPanel())
	b.WriteString("\n\n")

	hasFavs := len(m.favKeys) > 0
	if hasFavs {
		b.WriteString(m.st.Crumb.Render("favorites") + "\n")
	} else {
		b.WriteString(m.st.Crumb.Render("popular") +
			m.st.Help.Render("   (no favorites yet — press ") + m.st.Key.Render("f") +
			m.st.Help.Render(" on any station to save it)") + "\n")
	}

	// header(2) + blank(1) + panel(5) + blank(1) + label(1) + footer(1)
	listRows := m.height - 11
	if listRows < 3 {
		listRows = 3
	}
	if m.loading && len(m.stations) == 0 {
		b.WriteString(m.st.Meta.Render("  " + m.sp.View() + " loading…") + "\n")
	} else if len(m.stations) == 0 {
		b.WriteString(m.st.Meta.Render("  nothing to show") + "\n")
	} else {
		start, end := windowBounds(m.cursor, len(m.stations), listRows)
		for i := start; i < end; i++ {
			b.WriteString(m.renderRow(m.width, i, m.stations[i]) + "\n")
		}
	}
	b.WriteString("\n")
	b.WriteString(m.homeFooter())
	return b.String()
}

func (m Model) homeFooter() string {
	pairs := [][2]string{
		{"↑↓", "move"}, {"⏎", "play"}, {"+/-", "vol"}, {"[ ]", "quality"},
		{"/", "search"}, {"p", "popular"}, {"F", "favorites"}, {"a", "add"},
		{",", "settings"}, {"q", "quit"},
	}
	return m.renderFooterPairs(pairs)
}

// header renders the title line plus a right-aligned view label and a rule.
func (m Model) header(crumb string) string {
	left := m.st.Title.Render("onda") + m.st.Subtitle.Render("  ·  wander the airwaves")
	right := m.st.Crumb.Render(crumb)
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	rule := m.st.Rule.Render(strings.Repeat("─", max(1, m.width)))
	return left + strings.Repeat(" ", gap) + right + "\n" + rule
}

// updateBanner is a one-line notice shown under the header when a newer onda
// release exists. Its text branches on how onda was installed.
func (m Model) updateBanner() string {
	if !m.update.Available || m.updateDismiss {
		return ""
	}
	v := m.update.Latest
	var msg string
	switch {
	case m.updateApplying:
		msg = "updating to " + v + "…"
	case m.update.SelfUpdatable:
		msg = v + " available — press u to update"
	case m.update.InstallKind == "homebrew":
		msg = v + " available — run `brew upgrade --cask onda`"
	case m.update.InstallKind == "scoop":
		msg = v + " available — run `scoop update onda`"
	default:
		msg = v + " available — see github.com/pedrosousa13/onda/releases"
	}
	return m.st.Meta.Render("  ▲ "+msg) + m.st.Help.Render("  (U dismiss)")
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

// nowPanel is the bordered "now playing" hero. Line 1: station + volume,
// line 2: song / tags / status, line 3: the bitrate chooser.
func (m Model) nowPanel() string {
	inner := m.width - 4 // border(2) + padding(2)
	if inner < 12 {
		inner = 12
	}
	vol := m.volumeBar()

	// Line 1 — station name (left) + volume meter (right).
	var name string
	if m.isPlaying {
		name = m.st.NowTitle.Render("♫ " + truncate(m.playing.Name, max(4, inner-lipgloss.Width(vol)-1)))
	} else {
		name = m.st.NowTitle.Render("♫ nothing playing")
	}
	g1 := inner - lipgloss.Width(name) - lipgloss.Width(vol)
	if g1 < 1 {
		g1 = 1
	}
	line1 := name + strings.Repeat(" ", g1) + vol

	// Line 2 — current song (sanitized), else tags, else status.
	var line2 string
	switch {
	case !m.isPlaying:
		line2 = m.st.Meta.Render("select a station and press enter to play")
	case m.nowTitle != "":
		line2 = m.st.NowText.Render(truncate(m.nowTitle, inner))
	case len(m.playing.Tags) > 0:
		line2 = m.st.Meta.Render(truncate(strings.Join(m.playing.Tags, ", "), inner))
	default:
		line2 = m.st.Meta.Render("live")
	}

	// Line 3 — bitrate chooser (only when there's a choice).
	line3 := ""
	if m.isPlaying && len(m.playing.Variants) > 1 {
		line3 = m.qualityChips(inner)
	}

	content := line1 + "\n" + line2 + "\n" + line3
	// Panel.Width is the content box incl. padding(0,1); total = width-2+border(2) = width.
	return m.st.Panel.Width(m.width - 2).Render(content)
}

// qualityChips renders the playing station's available qualities; the active one
// is bracketed in the accent color so it's unmistakable.
func (m Model) qualityChips(maxW int) string {
	out := m.st.Help.Render("quality ")
	for i, v := range m.playing.Variants {
		var chip string
		if i == m.varIdx {
			chip = m.st.Crumb.Render("[" + v.Quality() + "]")
		} else {
			chip = m.st.Meta.Render(" " + v.Quality() + " ")
		}
		if lipgloss.Width(out+chip) > maxW {
			break
		}
		out += chip
	}
	return out
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
		{"a", "add"}, {",", "settings"}, {"esc", "home"}, {"q", "quit"},
	}
	return m.renderFooterPairs(pairs)
}

// renderFooterPairs lays out key/label hints, dropping trailing ones that don't fit.
func (m Model) renderFooterPairs(pairs [][2]string) string {
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
