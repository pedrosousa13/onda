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

// selectAsset returns the download URL of the archive matching goos+goarch,
// or "" if none. Matches by substring so a future name_template change is fine.
func selectAsset(assets []Asset, goos, goarch string) string {
	for _, a := range assets {
		n := strings.ToLower(a.Name)
		if isArchive(n) && strings.Contains(n, goos) && strings.Contains(n, goarch) {
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
