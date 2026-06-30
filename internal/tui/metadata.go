package tui

import (
	"regexp"
	"strings"
)

// Some streams push XML as ICY metadata (e.g. Dalet's <RadioInfo>…). These
// patterns pull a sensible "Artist — Title" out of the common cases.
var (
	daletTitle  = regexp.MustCompile(`(?is)<DB_DALET_TITLE_NAME>(.*?)</DB_DALET_TITLE_NAME>`)
	daletArtist = regexp.MustCompile(`(?is)<DB_DALET_ARTIST_NAME>(.*?)</DB_DALET_ARTIST_NAME>`)
	genericTtl = regexp.MustCompile(`(?is)<(?:title|song|track)>(.*?)</(?:title|song|track)>`)
)

// sanitizeTitle cleans ICY "now playing" metadata. Plain titles pass through;
// XML/markup is reduced to "Artist — Title" when recognized, else dropped
// (returns "") so the UI shows "live" rather than a raw XML blob.
func sanitizeTitle(s string) string {
	s = strings.TrimSpace(s)
	if !strings.Contains(s, "<") {
		return s
	}
	if t := firstGroup(daletTitle, s); t != "" {
		if a := firstGroup(daletArtist, s); a != "" {
			return a + " — " + t
		}
		return t
	}
	if t := firstGroup(genericTtl, s); t != "" {
		return t
	}
	return "" // unknown markup → show "live"
}

func firstGroup(re *regexp.Regexp, s string) string {
	m := re.FindStringSubmatch(s)
	if len(m) < 2 {
		return ""
	}
	return strings.TrimSpace(m[1])
}
