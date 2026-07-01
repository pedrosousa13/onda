package tui

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/pedrosousa13/onda/internal/directory"
	"github.com/pedrosousa13/onda/internal/domain"
)

// reserved vertical space: header(2) + blank(1) + blank(1) + now-panel(5) + footer(1).
const chromeHeight = 10

// gutter is the left/right breathing-room margin applied to every view.
const gutter = 2

// contentWidth is the usable width inside the left and right gutters.
func (m Model) contentWidth() int {
	w := m.width - 2*gutter
	if w < 20 {
		w = 20
	}
	return w
}

// indentLines prefixes every non-empty line with n spaces (the left gutter).
func indentLines(s string, n int) string {
	pad := strings.Repeat(" ", n)
	lines := strings.Split(s, "\n")
	for i, ln := range lines {
		if ln != "" {
			lines[i] = pad + ln
		}
	}
	return strings.Join(lines, "\n")
}

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
		b.WriteString("\n")
		b.WriteString(m.nowPanel(m.contentWidth()))
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
			b.WriteString(m.renderRow(m.contentWidth(), i, m.stations[i]) + "\n")
		}
		for i := end - start; i < listRows; i++ {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(m.nowPanel(m.contentWidth()))
	b.WriteString("\n")
	b.WriteString(m.footer())
	return b.String()
}

// viewBrowseMenu renders the browse axis/facet chooser: axis picker at level
// 0, then the facet list ("countries"/"genres"/"languages") at level 1.
func (m Model) viewBrowseMenu() string {
	crumb := "browse"
	if m.browseLevel != 0 {
		switch m.browseAxis {
		case domain.AxisTag:
			crumb = "genres"
		case domain.AxisLanguage:
			crumb = "languages"
		default:
			crumb = "countries"
		}
	}
	if m.loading {
		crumb = m.sp.View() + " loading"
	}

	var b strings.Builder
	b.WriteString(m.header(crumb))
	b.WriteString("\n\n")

	reserved := chromeHeight
	listRows := m.height - reserved
	if listRows < 3 {
		listRows = 3
	}

	if m.loading && len(m.facets) == 0 {
		b.WriteString(m.st.Meta.Render("  "+m.sp.View()+" finding facets…") + "\n")
		for i := 1; i < listRows; i++ {
			b.WriteString("\n")
		}
	} else {
		start, end := windowBounds(m.cursor, len(m.facets), listRows)
		for i := start; i < end; i++ {
			b.WriteString(m.renderFacetRow(m.contentWidth(), i, m.facets[i]) + "\n")
		}
		for i := end - start; i < listRows; i++ {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(m.nowPanel(m.contentWidth()))
	b.WriteString("\n")
	b.WriteString(m.browseMenuFooter())
	return b.String()
}

// browseMenuFooter is the key-hint bar for the browse axis/facet chooser.
func (m Model) browseMenuFooter() string {
	pairs := [][2]string{
		{"↑↓", "move"}, {"⏎", "open"}, {"esc", "back"}, {"q", "quit"},
	}
	return m.renderFooterPairs(pairs)
}

// viewHome is the landing screen: now-playing hero on top, then favorites
// (or a Popular preview when there are none).
func (m Model) viewHome() string {
	var b strings.Builder
	b.WriteString(m.header("home"))
	b.WriteString("\n")
	if m.bannerVisible() {
		b.WriteString(m.catalogBanner())
	}
	b.WriteString("\n")

	// Centered hero, capped so it doesn't stretch on wide terminals.
	heroWidth := m.contentWidth()
	if heroWidth > 56 {
		heroWidth = 56
	}
	b.WriteString(lipgloss.PlaceHorizontal(m.contentWidth(), lipgloss.Center, m.nowPanel(heroWidth)))
	b.WriteString("\n")
	hint := m.st.Help.Render("press ") + m.st.Key.Render("/") + m.st.Help.Render(" to search")
	b.WriteString(lipgloss.PlaceHorizontal(m.contentWidth(), lipgloss.Center, hint))
	b.WriteString("\n\n")

	cw := m.contentWidth()
	hasFavs := len(m.favKeys) > 0
	favLabel := func() {
		if hasFavs {
			b.WriteString(m.st.Crumb.Render("favorites") + "\n")
		} else {
			b.WriteString(m.st.Crumb.Render("popular") +
				m.st.Help.Render("   (no favorites yet — press ") + m.st.Key.Render("f") +
				m.st.Help.Render(" on any station to save it)") + "\n")
		}
	}

	recN := m.homeRecentsN()
	if recN == 0 {
		favLabel()
		// header(2) + blank(1) + panel(5) + hint(1) + blank(1) + label(1) + footer(1)
		listRows := m.height - 13
		if listRows < 3 {
			listRows = 3
		}
		if m.loading && len(m.stations) == 0 {
			b.WriteString(m.st.Meta.Render("  "+m.sp.View()+" loading…") + "\n")
		} else if len(m.stations) == 0 {
			b.WriteString(m.st.Meta.Render("  nothing to show") + "\n")
		} else {
			start, end := windowBounds(m.cursor, len(m.stations), listRows)
			for i := start; i < end; i++ {
				b.WriteString(m.renderRow(cw, i, m.stations[i]) + "\n")
			}
		}
		b.WriteString("\n")
		b.WriteString(m.homeFooter())
		return b.String()
	}

	// Two sections: pinned "recent" on top, then the scrolling favorites/popular
	// list. One extra label line vs the single-section layout → budget height-14.
	listRows := m.height - 14
	if listRows < 3 {
		listRows = 3
	}
	dispRecN, favStart, favEnd, _ := m.homeFavWindow(listRows)

	b.WriteString(m.st.Crumb.Render("recent") + "\n")
	for i := 0; i < dispRecN; i++ {
		b.WriteString(m.renderRow(cw, i, m.stations[i]) + "\n")
	}
	favLabel()
	if m.loading && favEnd == favStart {
		b.WriteString(m.st.Meta.Render("  "+m.sp.View()+" loading…") + "\n")
	} else {
		for j := favStart; j < favEnd; j++ {
			idx := recN + j
			b.WriteString(m.renderRow(cw, idx, m.stations[idx]) + "\n")
		}
	}
	b.WriteString("\n")
	b.WriteString(m.homeFooter())
	return b.String()
}

// homeFavWindow computes the Home two-section geometry: how many recent rows to
// pin (dispRecN, clamped to leave room) and the visible [favStart,favEnd) slice
// into the favorites/popular sub-list m.stations[recN:], keeping the cursor in view.
func (m Model) homeFavWindow(listRows int) (dispRecN, favStart, favEnd, favRows int) {
	recN := m.homeRecentsN()
	dispRecN = recN
	if dispRecN > listRows-1 {
		dispRecN = listRows - 1
	}
	if dispRecN < 0 {
		dispRecN = 0
	}
	favRows = listRows - dispRecN
	if favRows < 1 {
		favRows = 1
	}
	favTotal := len(m.stations) - recN
	if favTotal < 0 {
		favTotal = 0
	}
	favCursor := m.cursor - recN
	if favCursor < 0 {
		favCursor = 0
	}
	favStart, favEnd = windowBounds(favCursor, favTotal, favRows)
	return
}

func (m Model) homeFooter() string {
	pairs := [][2]string{
		{"↑↓", "move"}, {"⏎", "play"}, {"+/-", "vol"}, {"[ ]", "quality"},
		{"/", "search"}, {"b", "browse"}, {"p", "popular"}, {"F", "favorites"}, {"a", "add"},
		{",", "settings"}, {"q", "quit"},
	}
	return m.renderFooterPairs(pairs)
}

// header renders the title line plus a right-aligned view label and a rule.
func (m Model) header(crumb string) string {
	left := m.st.Title.Render("onda") + m.st.Subtitle.Render("  ·  wander the airwaves")
	right := m.st.Crumb.Render(crumb)
	if m.refreshing {
		mb := directory.HumanBytes(m.downloaded)
		right = m.st.Meta.Render(m.sp.View()+" building offline catalog… "+mb+"  ") + right
	}
	w := m.contentWidth()
	gap := w - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	rule := m.st.Rule.Render(strings.Repeat("─", max(1, w)))
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

// catalogBanner offers the full offline catalog on first launch, while
// consent is still undecided (offlineCatalog == "ask"). Shown only on Home.
func (m Model) catalogBanner() string {
	line1 := m.st.Meta.Render("  ⓘ Enable full offline catalog for typo-tolerant search?")
	line2 := m.st.Help.Render("    "+catalogSizeHint+", downloads in background.   ") +
		m.st.Key.Render("[y]") + m.st.Help.Render(" yes   ") +
		m.st.Key.Render("[n]") + m.st.Help.Render(" not now")
	return line1 + "\n" + line2 + "\n"
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
	if s.Votes > 0 {
		meta += " · " + humanCount(s.Votes) + "♥"
	}
	if s.Trend > 0 {
		meta += " ↑"
	}
	fav := m.favKeys[favKey(s)]
	starPlain := ""
	if fav {
		starPlain = " ★"
	}
	rightPlain := meta + starPlain

	avail := w - 2 /*marker*/ - 1 /*gap*/ - dispWidth(rightPlain)
	if avail < 4 {
		avail = 4
	}
	name := truncate(sanitizeName(s.Name), avail)
	pad := avail - dispWidth(name)
	if pad < 0 {
		pad = 0
	}

	var marker, nameS string
	switch {
	case sel:
		marker = m.st.SelBar.Render("▌ ")
		nameS = m.st.SelName.Render(name)
	case idx == m.hoverIdx:
		marker = m.st.Meta.Render("· ") // mouse hover cue
		nameS = m.st.Item.Render(name)
	default:
		marker = "  "
		nameS = m.st.Item.Render(name)
	}
	starS := ""
	if fav {
		starS = " " + m.st.Star.Render("★")
	}
	return marker + nameS + strings.Repeat(" ", pad) + " " + m.st.Meta.Render(meta) + starS
}

// humanCount formats a count for compact display: 412→"412", 1200→"1.2k".
func humanCount(n int) string {
	if n < 1000 {
		return strconv.Itoa(n)
	}
	return strconv.FormatFloat(float64(n)/1000, 'f', 1, 64) + "k"
}

// renderFacetRow lays out one browse facet: ▌ Portugal … 412
func (m Model) renderFacetRow(w, idx int, f domain.Facet) string {
	sel := idx == m.cursor

	rightPlain := ""
	if f.Count > 0 {
		rightPlain = humanCount(f.Count)
	}

	avail := w - 2 /*marker*/ - 1 /*gap*/ - dispWidth(rightPlain)
	if avail < 4 {
		avail = 4
	}
	name := truncate(sanitizeName(f.Name), avail)
	pad := avail - dispWidth(name)
	if pad < 0 {
		pad = 0
	}

	var marker, nameS string
	switch {
	case sel:
		marker = m.st.SelBar.Render("▌ ")
		nameS = m.st.SelName.Render(name)
	case idx == m.hoverIdx:
		marker = m.st.Meta.Render("· ") // mouse hover cue
		nameS = m.st.Item.Render(name)
	default:
		marker = "  "
		nameS = m.st.Item.Render(name)
	}
	right := ""
	if f.Count > 0 {
		right = " " + m.st.Meta.Render(rightPlain)
	}
	return marker + nameS + strings.Repeat(" ", pad) + right
}

// nowPanel is the bordered "now playing" hero. Line 1: station + volume,
// line 2: song / tags / status, line 3: the bitrate chooser.
func (m Model) nowPanel(width int) string {
	inner := width - 4 // border(2) + padding(2)
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

	// Line 2 — phase-aware: connecting / error, else song, tags, or status.
	var line2 string
	switch {
	case m.phase == phaseFailed:
		line2 = m.st.Subtitle.Render(truncate(m.playErr, inner))
	case m.phase == phaseConnecting:
		line2 = m.st.NowTitle.Render(truncate("connecting…", inner))
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
	return m.st.Panel.Width(width - 2).Render(content)
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
	escPair := [2]string{"esc", "home"}
	if m.browseLevel == 2 {
		escPair = [2]string{"esc", "back"}
	}
	pairs := [][2]string{
		{"↑↓", "move"}, {"⏎", "play"}, {"s", "stop"}, {"+/-", "vol"},
		{"[ ]", "quality"}, {"f", "★"}, {"F", "favs"}, {"/", "search"},
		{"b", "browse"}, {"a", "add"}, {",", "settings"}, escPair, {"q", "quit"},
	}
	if m.browseLevel == 2 {
		pairs = append(pairs, [2]string{"o", "sort"}, [2]string{"O", "reverse"})
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
		if wsum+add > m.contentWidth() {
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

// clampWidth hard-caps every line of s to w display columns, so a stray
// over-wide line can never soft-wrap in the terminal and desync Bubble Tea's
// line-diff renderer. This is the single authoritative backstop on top of
// per-row truncation.
//
// It measures conservatively (dispWidth) rather than with lipgloss's
// Unicode-correct width: a complex-script line (Tamil, Devanagari, …) can be
// DRAWN wider than lipgloss reports, so lipgloss.MaxWidth would let it slip
// through — exactly the wrap that desyncs the renderer. Over-cutting such a
// line is the safe direction.
func clampWidth(s string, w int) string {
	if w <= 0 {
		return s
	}
	lines := strings.Split(s, "\n")
	for i, ln := range lines {
		lines[i] = clampLine(ln, w)
	}
	return strings.Join(lines, "\n")
}

// clampLine truncates one line to at most w display columns, measured with the
// conservative dispWidth and skipping ANSI SGR escape sequences (which occupy
// no columns). If it has to cut, it appends a reset so a severed styled span
// can't bleed color into the rest of the frame.
func clampLine(s string, w int) string {
	var b strings.Builder
	width := 0
	runes := []rune(s)
	cut := false
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == 0x1b { // ESC — copy the whole escape sequence verbatim, uncounted
			j := i + 1
			if j < len(runes) && runes[j] == '[' { // CSI: ESC '[' … final byte 0x40–0x7e
				j++
				for j < len(runes) && (runes[j] < 0x40 || runes[j] > 0x7e) {
					j++
				}
				if j < len(runes) {
					j++ // include the final byte
				}
			}
			for k := i; k < j; k++ {
				b.WriteRune(runes[k])
			}
			i = j - 1
			continue
		}
		rw := runeCells(r)
		if width+rw > w {
			cut = true
			break
		}
		b.WriteRune(r)
		width += rw
	}
	if cut {
		b.WriteString("\x1b[0m")
	}
	return b.String()
}

// runeCells is a CONSERVATIVE cell count for one rune: wide glyphs (CJK) keep
// their 2 cells, but everything else — including combining marks that Unicode
// (and lipgloss) count as 0 — is treated as at least 1 cell. Terminals that
// don't shape complex scripts (Tamil, Devanagari, …) render each codepoint in
// its own cell, so lipgloss's Unicode-correct width UNDER-counts what the
// terminal actually draws. Over-counting here is the safe direction: a row can
// never end up wider than its budget, which would soft-wrap and desync Bubble
// Tea's line-diff renderer.
func runeCells(r rune) int {
	if w := lipgloss.Width(string(r)); w > 1 {
		return w
	}
	// Complex scripts (Arabic, Indic, SE-Asian — e.g. Tamil) are drawn with
	// combining marks and base glyphs that many terminals/fonts render WIDER
	// than lipgloss's Unicode width reports (verified: a standard grid fits a
	// Tamil row that a real terminal wraps). A monospace cell can hold at most
	// 2 columns per codepoint, so reserving 2 here upper-bounds whatever the
	// terminal actually draws and guarantees the row can't overflow and wrap.
	if r >= 0x0600 && r < 0x1100 {
		return 2
	}
	return 1
}

// dispWidth is the conservative display width of s (see runeCells).
func dispWidth(s string) int {
	w := 0
	for _, r := range s {
		w += runeCells(r)
	}
	return w
}

// sanitizeName strips control characters (tabs, stray newlines from dirty
// station data) that would otherwise break row layout, and trims the result.
func sanitizeName(s string) string {
	s = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return ' '
		}
		return r
	}, s)
	return strings.TrimSpace(s)
}

// truncate shortens s to at most w display columns, adding an ellipsis.
// It measures conservatively (dispWidth) so wide or complex-script glyphs can't
// produce a row wider than its budget — an overflowing row soft-wraps in the
// terminal and desyncs Bubble Tea's line-diff renderer.
func truncate(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if dispWidth(s) <= w {
		return s
	}
	if w == 1 {
		return "…"
	}
	var b strings.Builder
	width := 0
	for _, r := range s {
		rw := runeCells(r)
		if width+rw > w-1 { // leave one column for the ellipsis
			break
		}
		b.WriteRune(r)
		width += rw
	}
	return b.String() + "…"
}
