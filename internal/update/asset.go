package update

import "strings"

// Asset is a release asset (subset of the GitHub API shape).
type Asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

// isArchive tie-breaks against non-archive assets (sbom, sig) that may share
// the os/arch tokens.
func isArchive(name string) bool {
	return strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".zip")
}

// archMatches checks for the arch as a delimited token, not a raw substring, so
// goarch "arm" does not match an "..._arm64..." asset (GoReleaser names archs
// between "_" and the extension/next field).
func archMatches(name, goarch string) bool {
	return strings.Contains(name, "_"+goarch+".") || strings.Contains(name, "_"+goarch+"_")
}

// selectAsset returns the download URL of the archive matching goos+goarch,
// or "" if none. Matches by substring (os) / delimited token (arch) so a future
// name_template change is tolerated while avoiding the arm⊂arm64 false match.
func selectAsset(assets []Asset, goos, goarch string) string {
	for _, a := range assets {
		n := strings.ToLower(a.Name)
		if isArchive(n) && strings.Contains(n, goos) && archMatches(n, goarch) {
			return a.URL
		}
	}
	return ""
}

// checksumsURL returns the checksums.txt asset URL, or "".
func checksumsURL(assets []Asset) string {
	for _, a := range assets {
		if a.Name == "checksums.txt" {
			return a.URL
		}
	}
	return ""
}
