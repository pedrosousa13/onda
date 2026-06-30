package update

import (
	"testing"
	"time"
)

func TestCacheRoundTripAndTTL(t *testing.T) {
	dir := t.TempDir()
	c := cache{LatestVersion: "v1.2.0", AssetURL: "a", ChecksumsURL: "s", CheckedAt: time.Unix(1000, 0)}
	if err := saveCache(dir, c); err != nil {
		t.Fatal(err)
	}
	got, ok := loadCache(dir)
	if !ok || got.LatestVersion != "v1.2.0" || got.AssetURL != "a" {
		t.Fatalf("round-trip failed: %+v ok=%v", got, ok)
	}

	now := time.Unix(1000, 0)
	if !fresh(c, now.Add(23*time.Hour)) {
		t.Fatal("23h should be fresh")
	}
	if fresh(c, now.Add(25*time.Hour)) {
		t.Fatal("25h should be stale")
	}
}

func TestLoadCacheMissing(t *testing.T) {
	if _, ok := loadCache(t.TempDir()); ok {
		t.Fatal("missing cache should report ok=false")
	}
}
