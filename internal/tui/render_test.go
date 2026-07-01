package tui

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/pedrosousa13/onda/internal/domain"
	"github.com/pedrosousa13/onda/internal/update"
)

var ansiRe = regexp.MustCompile("\x1b\\[[0-9;]*m")

func stripANSI(s string) string { return ansiRe.ReplaceAllString(s, "") }

// sampleModel builds a populated model for visual inspection of View() output.
func sampleModel() Model {
	m := New(nil, nil, nil, domain.QualityHighest, "never", false, "catppuccin-mocha", true, true, 100, false, false, "ask", "1.0.0", "")
	m.width, m.height = 76, 22
	m.stations = []domain.Station{
		{Name: "KEXP", Country: "United States", Tags: []string{"indie", "seattle"}, Variants: []domain.StreamVariant{{Bitrate: 128}}},
		{Name: "BBC World Service", Country: "United Kingdom", Tags: []string{"news"}, Variants: []domain.StreamVariant{{Bitrate: 96}}},
		{Name: "FIP", Country: "France", Tags: []string{"eclectic", "jazz"}, Variants: []domain.StreamVariant{{URL: "h", Lossless: true}, {URL: "a", Bitrate: 192}, {URL: "b", Bitrate: 128}}},
		{Name: "NTS Radio 1", Country: "United Kingdom", Tags: []string{"electronic"}, Variants: []domain.StreamVariant{{Bitrate: 128}}},
		{Name: "Radio Nacional de España", Country: "Spain", Variants: []domain.StreamVariant{{Bitrate: 64}}},
		{Name: "Triple J", Country: "Australia", Tags: []string{"alternative"}, Variants: []domain.StreamVariant{{Bitrate: 128}}},
	}
	m.cursor = 2
	m.loading = false
	m.crumb = "popular"
	m.favKeys = map[string]bool{favKey(m.stations[0]): true}
	m.playing = m.stations[2]
	m.isPlaying = true
	m.nowTitle = "Khruangbin — Maria También"
	m.volume = 80
	return m
}

// TestRenderGallery prints each screen so the layout can be reviewed.
// Run: go test ./internal/tui/ -run TestRenderGallery -v
func TestRenderGallery(t *testing.T) {
	frame := func(label string, m Model) {
		fmt.Printf("\n┌──────── %s ", label)
		for i := len(label); i < 60; i++ {
			fmt.Print("─")
		}
		fmt.Println("┐")
		fmt.Println(m.View())
		fmt.Println("└" + repeat("─", 70) + "┘")
	}

	m := sampleModel()
	frame("HOME", m) // sampleModel has a favorite, so Home shows favorites

	bw := m
	bw.view = viewBrowse
	bw.crumb = "popular"
	frame("BROWSE", bw)

	s := m
	s.view = viewSearch
	s.search.SetValue("jazz")
	frame("SEARCH", s)

	a := m
	a.view = viewAdd
	a.addName.SetValue("My Stream")
	a.addURL.SetValue("https://example.com/stream")
	frame("ADD STATION", a)

	st := m
	st.view = viewSettings
	frame("SETTINGS", st)

	f := m
	f.view = viewFavorites
	frame("FAVORITES", f)
}

// TestGutterPadding verifies every view's non-empty lines sit inside the left
// gutter, and that content is laid out at contentWidth (leaving a right gutter).
func TestGutterPadding(t *testing.T) {
	pad := strings.Repeat(" ", gutter)
	for _, v := range []view{viewBrowse, viewHome, viewSettings, viewSearch, viewAdd} {
		m := sampleModel()
		m.view = v
		for _, ln := range strings.Split(m.View(), "\n") {
			if ln == "" {
				continue
			}
			if !strings.HasPrefix(ln, pad) {
				t.Errorf("view %d: line not gutter-indented: %q", v, ln)
			}
		}
	}
}

// TestPanelWidthWithinGutter checks the now-panel renders at the content width,
// so the indented output never exceeds the terminal width.
func TestPanelWidthWithinGutter(t *testing.T) {
	m := sampleModel()
	panel := m.nowPanel(m.contentWidth())
	for _, ln := range strings.Split(panel, "\n") {
		if w := lipgloss.Width(ln); w > m.contentWidth() {
			t.Errorf("panel line width %d exceeds contentWidth %d: %q", w, m.contentWidth(), ln)
		}
	}
	if got := m.contentWidth(); got != m.width-2*gutter {
		t.Errorf("contentWidth = %d, want %d", got, m.width-2*gutter)
	}
}

// TestHomeCenteredHeroAndHint verifies the Home hero is centered (indented
// beyond the gutter on a wide terminal) and the search hint is present.
func TestHomeCenteredHeroAndHint(t *testing.T) {
	m := sampleModel()
	m.view = viewHome
	out := m.View()
	if !strings.Contains(out, "to search") {
		t.Error("home view missing search hint")
	}
	var centered bool
	for _, ln := range strings.Split(out, "\n") {
		if strings.Contains(ln, "╭") {
			lead := len(ln) - len(strings.TrimLeft(ln, " "))
			centered = lead > gutter
			break
		}
	}
	if !centered {
		t.Error("hero panel not centered beyond the gutter")
	}
}

func TestUpdateBannerText(t *testing.T) {
	mk := func(s update.Status) Model {
		return Model{st: newStyles(themeByName("catppuccin-mocha")), width: 80, update: s}
	}
	cases := []struct{ kind, want string }{
		{"homebrew", "brew upgrade"},
		{"scoop", "scoop update"},
		{"unknown", "releases"},
	}
	for _, c := range cases {
		m := mk(update.Status{Available: true, Latest: "v2.0.0", InstallKind: c.kind})
		if got := m.updateBanner(); !strings.Contains(got, c.want) {
			t.Errorf("kind %q banner = %q, want contains %q", c.kind, got, c.want)
		}
	}
	if got := mk(update.Status{Available: true, Latest: "v2.0.0", InstallKind: "binary", SelfUpdatable: true}).updateBanner(); !strings.Contains(got, "press u") {
		t.Errorf("self-updatable banner = %q, want key hint", got)
	}
	if got := (Model{}).updateBanner(); got != "" {
		t.Errorf("no update should render empty banner, got %q", got)
	}
	// Dismissed banner renders empty even when an update is available.
	dismissed := mk(update.Status{Available: true, Latest: "v2.0.0", InstallKind: "homebrew"})
	dismissed.updateDismiss = true
	if got := dismissed.updateBanner(); got != "" {
		t.Errorf("dismissed banner should be empty, got %q", got)
	}
}

func repeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}

func TestHumanCount(t *testing.T) {
	cases := []struct {
		in   int
		want string
	}{
		{412, "412"},
		{1200, "1.2k"},
		{0, "0"},
		{999, "999"},
	}
	for _, c := range cases {
		if got := humanCount(c.in); got != c.want {
			t.Errorf("humanCount(%d) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestRenderRowShowsVotesTrendOnlyWhenPresent(t *testing.T) {
	m := Model{width: 60, st: newStyles(themeByName("catppuccin-mocha")), favKeys: map[string]bool{}, hoverIdx: -1}

	withBoth := domain.Station{Name: "KEXP", Country: "United States", Votes: 1200, Trend: 1}
	row := m.renderRow(m.contentWidth(), 0, withBoth)
	if !strings.Contains(row, "1.2k♥") {
		t.Errorf("row with votes = %q, want contains %q", row, "1.2k♥")
	}
	if !strings.Contains(row, "↑") {
		t.Errorf("row with trend = %q, want contains %q", row, "↑")
	}

	bare := domain.Station{Name: "KEXP", Country: "United States"}
	row = m.renderRow(m.contentWidth(), 0, bare)
	if strings.Contains(row, "♥") {
		t.Errorf("bare row = %q, want no %q", row, "♥")
	}
	if strings.Contains(row, "↑") {
		t.Errorf("bare row = %q, want no %q", row, "↑")
	}
}

// wideName mixes 2-column CJK glyphs so display width exceeds rune count —
// the case that made rows soft-wrap and desync the renderer.
const wideName = "日本語ラジオ放送局とても長い名前のテスト局です"

func TestTruncateBoundsDisplayWidth(t *testing.T) {
	if got := lipgloss.Width(truncate(wideName, 10)); got > 10 {
		t.Fatalf("truncate width = %d, want <= 10", got)
	}
	if got := truncate("hello world", 100); got != "hello world" {
		t.Fatalf("short ASCII should pass through, got %q", got)
	}
	if got := truncate("hello world", 5); lipgloss.Width(got) > 5 {
		t.Fatalf("ASCII truncate width = %d, want <= 5", lipgloss.Width(got))
	}
}

func TestRenderRowWideNameNeverExceedsWidth(t *testing.T) {
	m := sampleModel()
	s := domain.Station{Name: wideName, Country: "Japan", Votes: 1200, Trend: 3}
	row := m.renderRow(m.contentWidth(), 0, s)
	if w := lipgloss.Width(row); w > m.contentWidth() {
		t.Fatalf("renderRow width = %d, want <= contentWidth %d", w, m.contentWidth())
	}
}

func TestRenderFacetRowWideNameNeverExceedsWidth(t *testing.T) {
	m := sampleModel()
	f := domain.Facet{Name: wideName, Count: 1234}
	row := m.renderFacetRow(m.contentWidth(), 0, f)
	if w := lipgloss.Width(row); w > m.contentWidth() {
		t.Fatalf("renderFacetRow width = %d, want <= contentWidth %d", w, m.contentWidth())
	}
}

// TestViewNeverExceedsTerminalWidth is the belt-and-suspenders guarantee: no
// matter how wide a line's content is (long meta on a narrow terminal, dirty
// data, complex-script crumb or station name), View() must never emit a line
// wider than the terminal — a wrapped line desyncs Bubble Tea's renderer.
// Measured with the conservative dispWidth (what the terminal may actually
// draw), NOT lipgloss.Width — lipgloss under-counts complex scripts, which is
// the whole bug: a line lipgloss thinks fits still overflows and wraps.
func TestViewNeverExceedsTerminalWidth(t *testing.T) {
	// A Tamil crumb flows into the header, and a Tamil station into the list —
	// both are lines that lipgloss under-measures and the backstop must cap.
	const tamilName = "Tube Tamil FM Radio டியூப் தமிழ் எஃப்.எம் பண்பலை ரேடியோ"
	for _, w := range []int{24, 40, 60, 100} {
		m := Model{
			width: w, height: 24, st: newStyles(themeByName("catppuccin-mocha")),
			favKeys: map[string]bool{}, hoverIdx: -1, view: viewBrowse,
			crumb: "தமிழ் மொழி · votes",
		}
		m.stations = []domain.Station{
			{Name: "Some Very Long Station Name Indeed", Country: "The United Kingdom Of Great Britain And Northern Ireland", Votes: 304300, Trend: 3},
			{Name: wideName, Country: "Japan", Votes: 1200, Trend: 5},
			{Name: tamilName, Country: "Sri Lanka", Votes: 4600},
		}
		for _, ln := range strings.Split(m.View(), "\n") {
			if dw := dispWidth(stripANSI(ln)); dw > m.width {
				t.Fatalf("width %d: View line conservative width %d exceeds terminal width %d: %q", w, dw, m.width, stripANSI(ln))
			}
		}
	}
}

func TestDispWidthConservative(t *testing.T) {
	if got := dispWidth("ab"); got != 2 {
		t.Fatalf("dispWidth(ab) = %d, want 2", got)
	}
	if got := dispWidth("日本語"); got != 6 { // 3 CJK glyphs, 2 cells each
		t.Fatalf("dispWidth(CJK) = %d, want 6", got)
	}
	// Complex scripts (Tamil) reserve the monospace max (2 cells/codepoint), so
	// dispWidth is at least the rune count and never below lipgloss's width —
	// the safe direction against terminals that render them wider than Unicode.
	tamil := "தமிழ்"
	if dispWidth(tamil) < utf8.RuneCountInString(tamil) {
		t.Fatalf("dispWidth(tamil) = %d, want >= rune count %d", dispWidth(tamil), utf8.RuneCountInString(tamil))
	}
	if dispWidth(tamil) < lipgloss.Width(tamil) {
		t.Fatalf("dispWidth must never undercount lipgloss: %d < %d", dispWidth(tamil), lipgloss.Width(tamil))
	}
	if got, want := dispWidth(tamil), 2*utf8.RuneCountInString(tamil); got != want {
		t.Fatalf("complex-script dispWidth = %d, want 2 cells/codepoint = %d", got, want)
	}
}

// TestClampWidthConservativeAndANSIAware verifies the View-level backstop caps
// lines by the terminal-real (dispWidth) measure, keeps ANSI escapes out of the
// count, and doesn't let a severed styled span bleed color.
func TestClampWidthConservativeAndANSIAware(t *testing.T) {
	// ANSI codes occupy no columns: a styled 5-char string survives a width-5 cap.
	styled := "\x1b[31mhello\x1b[0m"
	if got := clampWidth(styled, 5); got != styled {
		t.Fatalf("styled fit was altered: %q", got)
	}

	// A complex-script line lipgloss under-measures must be cut to <= w by dispWidth.
	tamil := "தமிழ் மொழி பண்பலை ரேடியோ ஒலிபரப்பு"
	if lipgloss.Width(tamil) <= 8 {
		// sanity: lipgloss thinks it fits — which is exactly why MaxWidth failed.
	}
	out := clampWidth(tamil, 8)
	if dw := dispWidth(stripANSI(out)); dw > 8 {
		t.Fatalf("clamped complex-script width = %d, want <= 8: %q", dw, stripANSI(out))
	}

	// Cutting a colored line appends a reset so color can't bleed downstream.
	if got := clampWidth("\x1b[31m"+strings.Repeat("x", 20), 5); !strings.HasSuffix(got, "\x1b[0m") {
		t.Fatalf("cut styled line missing trailing reset: %q", got)
	}

	// Per-line: each line of a multi-line block is capped independently.
	for _, ln := range strings.Split(clampWidth("aaaaaaaa\nbbbb\n"+tamil, 4), "\n") {
		if dw := dispWidth(stripANSI(ln)); dw > 4 {
			t.Fatalf("multiline clamp left width %d > 4: %q", dw, ln)
		}
	}
}

func TestSanitizeNameStripsControl(t *testing.T) {
	if got := sanitizeName("a\tb\nc"); strings.ContainsAny(got, "\t\n") {
		t.Fatalf("control chars not stripped: %q", got)
	}
	if got := sanitizeName("  padded  "); got != "padded" {
		t.Fatalf("sanitizeName should trim, got %q", got)
	}
}

// The real station + terminal width that reproduced the wrap/desync bug: a
// Tamil name lipgloss measured shorter than the terminal renders it. Measured
// conservatively (as the terminal draws it), the row must fit.
func TestRenderRowComplexScriptNeverOverflowsTerminal(t *testing.T) {
	m := Model{
		width: 60, height: 24, st: newStyles(themeByName("catppuccin-mocha")),
		favKeys: map[string]bool{}, hoverIdx: -1,
	}
	s := domain.Station{
		Name:    "Tube Tamil FM Radio டியூப் தமிழ் எஃப்.எம் பண்பலை ரேடியோ",
		Country: "Sri Lanka", Votes: 4600,
	}
	plain := stripANSI(m.renderRow(m.contentWidth(), 0, s))
	if got := dispWidth(plain); got > m.contentWidth() {
		t.Fatalf("row conservative width = %d, want <= contentWidth %d: %q", got, m.contentWidth(), plain)
	}
}
