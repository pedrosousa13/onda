package update

import "testing"

func TestSelectAsset(t *testing.T) {
	assets := []Asset{
		{Name: "onda_1.0.0_darwin_amd64.tar.gz", URL: "u-dar-amd"},
		{Name: "onda_1.0.0_darwin_arm64.tar.gz", URL: "u-dar-arm"},
		{Name: "onda_1.0.0_windows_amd64.zip", URL: "u-win-amd"},
		{Name: "checksums.txt", URL: "u-sums"},
		{Name: "onda_1.0.0_darwin_arm64.sbom.json", URL: "u-sbom"}, // must NOT win
	}
	if got := selectAsset(assets, "darwin", "arm64"); got != "u-dar-arm" {
		t.Errorf("darwin/arm64 -> %q", got)
	}
	if got := selectAsset(assets, "windows", "amd64"); got != "u-win-amd" {
		t.Errorf("windows/amd64 -> %q", got)
	}
	if got := selectAsset(assets, "linux", "arm64"); got != "" {
		t.Errorf("unsupported target should be empty, got %q", got)
	}
	// 32-bit arm must NOT false-match the arm64 archive (arm ⊂ arm64).
	if got := selectAsset(assets, "linux", "arm"); got != "" {
		t.Errorf("32-bit arm should not match arm64 asset, got %q", got)
	}
	if got := selectAsset(assets, "darwin", "arm"); got != "" {
		t.Errorf("darwin/arm should not match darwin/arm64, got %q", got)
	}
	if got := checksumsURL(assets); got != "u-sums" {
		t.Errorf("checksumsURL -> %q", got)
	}
}
