package directory

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/pedrosousa13/radio/internal/domain"
)

type Cache struct {
	dir string
	ttl time.Duration
	now func() time.Time
}

func NewCache(dir string, ttl time.Duration) *Cache {
	return &Cache{dir: dir, ttl: ttl, now: time.Now}
}

type cacheEntry struct {
	StoredAt time.Time        `json:"stored_at"`
	Stations []domain.Station `json:"stations"`
}

func (c *Cache) path(query string) string {
	sum := sha256.Sum256([]byte(query))
	return filepath.Join(c.dir, hex.EncodeToString(sum[:8])+".json")
}

func (c *Cache) read(query string) (cacheEntry, bool) {
	b, err := os.ReadFile(c.path(query))
	if err != nil {
		return cacheEntry{}, false
	}
	var e cacheEntry
	if json.Unmarshal(b, &e) != nil {
		return cacheEntry{}, false
	}
	return e, true
}

// Get returns cached stations only if still within TTL.
func (c *Cache) Get(query string) ([]domain.Station, bool) {
	e, ok := c.read(query)
	if !ok || c.now().Sub(e.StoredAt) > c.ttl {
		return nil, false
	}
	return e.Stations, true
}

// Stale returns last-known stations regardless of TTL (offline fallback).
func (c *Cache) Stale(query string) ([]domain.Station, bool) {
	e, ok := c.read(query)
	if !ok {
		return nil, false
	}
	return e.Stations, true
}

func (c *Cache) Put(query string, stations []domain.Station) error {
	if err := os.MkdirAll(c.dir, 0o755); err != nil {
		return err
	}
	b, err := json.Marshal(cacheEntry{StoredAt: c.now(), Stations: stations})
	if err != nil {
		return err
	}
	return os.WriteFile(c.path(query), b, 0o644)
}
