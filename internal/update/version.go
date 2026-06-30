package update

import (
	"strings"

	"golang.org/x/mod/semver"
)

// vprefix ensures exactly one leading "v" so x/mod/semver accepts it.
func vprefix(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if !strings.HasPrefix(s, "v") {
		return "v" + s
	}
	return s
}

// isNewer reports whether latest is a strictly greater valid semver than current.
func isNewer(current, latest string) bool {
	c, l := vprefix(current), vprefix(latest)
	if !semver.IsValid(c) || !semver.IsValid(l) {
		return false
	}
	return semver.Compare(c, l) < 0
}
