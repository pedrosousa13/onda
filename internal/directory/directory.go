// Package directory provides station data from multiple sources behind one interface.
package directory

import (
	"context"

	"github.com/pedrosousa13/radio/internal/domain"
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
