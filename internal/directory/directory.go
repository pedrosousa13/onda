// Package directory provides station data from multiple sources behind one interface.
package directory

import (
	"context"

	"github.com/pedrosousa13/onda/internal/domain"
)

// Source is a provider of stations (online or offline).
type Source interface {
	// Search returns stations matching query (empty query => a sensible default set).
	Search(ctx context.Context, query string) ([]domain.Station, error)
}

// Directory aggregates sources with caching and offline fallback.
type Directory struct {
	Online  Source
	Offline Source
	Cache   *Cache
}

func (d *Directory) Search(ctx context.Context, query string) ([]domain.Station, error) {
	if d.Cache != nil {
		if fresh, ok := d.Cache.Get(query); ok {
			return fresh, nil
		}
	}
	if d.Online != nil {
		if stations, err := d.Online.Search(ctx, query); err == nil {
			if d.Cache != nil {
				_ = d.Cache.Put(query, stations)
			}
			return stations, nil
		}
	}
	if d.Cache != nil {
		if stale, ok := d.Cache.Stale(query); ok {
			return stale, nil
		}
	}
	return d.Offline.Search(ctx, query)
}

// popularSource is optionally implemented by an online Source that can return
// a community-popularity ranking.
type popularSource interface {
	TopVoted(ctx context.Context, limit int) ([]domain.Station, error)
}

const popularKey = "__popular__"

// Popular returns the top-voted stations (read-only; reports nothing about the
// user). Falls back to cache, then the embedded offline list, when offline.
func (d *Directory) Popular(ctx context.Context) ([]domain.Station, error) {
	if d.Cache != nil {
		if fresh, ok := d.Cache.Get(popularKey); ok {
			return fresh, nil
		}
	}
	if ps, ok := d.Online.(popularSource); ok && d.Online != nil {
		if stations, err := ps.TopVoted(ctx, 100); err == nil {
			if d.Cache != nil {
				_ = d.Cache.Put(popularKey, stations)
			}
			return stations, nil
		}
	}
	if d.Cache != nil {
		if stale, ok := d.Cache.Stale(popularKey); ok {
			return stale, nil
		}
	}
	return d.Offline.Search(ctx, "")
}
