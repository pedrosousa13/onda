package update

import (
	"os"
	"path/filepath"
	"strings"
)

// classifyPath maps an executable path to "homebrew", "scoop", or "other".
// Conservative: anything not clearly a package manager is "other".
func classifyPath(exe string) string {
	// Normalize separators regardless of host OS so detection is consistent and
	// unit-testable (filepath.ToSlash is a no-op for backslashes off Windows).
	p := strings.ReplaceAll(exe, "\\", "/")
	low := strings.ToLower(p)
	switch {
	case strings.Contains(p, "/Cellar/"),
		strings.Contains(p, "/opt/homebrew/"),
		strings.Contains(p, "/.linuxbrew/"):
		return "homebrew"
	case strings.Contains(low, "/scoop/apps/"):
		return "scoop"
	default:
		return "other"
	}
}

// dirWritable probes by creating and removing a temp file (mode bits are
// unreliable on Windows). Advisory only — Apply does the authoritative write.
func dirWritable(dir string) bool {
	f, err := os.CreateTemp(dir, ".onda-write-probe-*")
	if err != nil {
		return false
	}
	name := f.Name()
	f.Close()
	os.Remove(name)
	return true
}

// installKind resolves the running binary's path (following symlinks) and
// classifies it. Returns ("unknown", false) if the path can't be resolved.
func installKind() (kind string, writable bool) {
	exe, err := os.Executable()
	if err != nil {
		return "unknown", false
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	kind = classifyPath(exe)
	if kind == "other" {
		if dirWritable(filepath.Dir(exe)) {
			return "binary", true
		}
		return "unknown", false
	}
	return kind, false
}
