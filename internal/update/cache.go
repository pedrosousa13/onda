package update

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const cacheTTL = 24 * time.Hour

type cache struct {
	CheckedAt     time.Time `json:"checked_at"`
	LatestVersion string    `json:"latest_version"`
	AssetURL      string    `json:"asset_url"`
	ChecksumsURL  string    `json:"checksums_url"`
}

func cachePath(dir string) string { return filepath.Join(dir, "update-cache.json") }

func loadCache(dir string) (cache, bool) {
	b, err := os.ReadFile(cachePath(dir))
	if err != nil {
		return cache{}, false
	}
	var c cache
	if json.Unmarshal(b, &c) != nil {
		return cache{}, false
	}
	return c, true
}

func saveCache(dir string, c cache) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(cachePath(dir), b, 0o644)
}

func fresh(c cache, now time.Time) bool {
	return now.Sub(c.CheckedAt) < cacheTTL
}
