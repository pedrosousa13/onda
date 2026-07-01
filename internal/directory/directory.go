// Package directory provides station data from multiple sources behind one interface.
package directory

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/pedrosousa13/onda/internal/domain"
)

// Source is a provider of stations (online or offline).
type Source interface {
	// Search returns stations matching query (empty query => a sensible default set).
	Search(ctx context.Context, query string) ([]domain.Station, error)
}

// Directory serves stations from an in-memory corpus (the full Radio Browser
// dump). The network is touched only by Refresh.
type Directory struct {
	Online  Source       // used only by Refresh (must implement fullFetcher)
	Offline Source       // embedded starter list, the cold-start/offline floor
	Corpus  *CorpusStore // on-disk persistence; nil disables persistence

	mu     sync.RWMutex
	corpus []domain.Station
}

type fullFetcher interface {
	FetchAll(ctx context.Context) ([]domain.Station, error)
}

func (d *Directory) setCorpus(s []domain.Station) {
	d.mu.Lock()
	d.corpus = s
	d.mu.Unlock()
}

func (d *Directory) snapshot() []domain.Station {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.corpus
}

// LoadCorpus loads any cached dump into memory. Returns true when the loaded
// dump is still within TTL (caller skips an immediate refresh).
func (d *Directory) LoadCorpus() bool {
	if d.Corpus == nil {
		return false
	}
	if st, at, ok := d.Corpus.Load(); ok {
		d.setCorpus(st)
		return d.Corpus.Fresh(at)
	}
	return false
}

// base returns the corpus, falling back to the embedded starter list.
func (d *Directory) base(ctx context.Context) ([]domain.Station, error) {
	if s := d.snapshot(); len(s) > 0 {
		return s, nil
	}
	if d.Offline != nil {
		return d.Offline.Search(ctx, "")
	}
	return nil, nil
}

// Initial is the cold-start list: the corpus if loaded, else the embedded list.
func (d *Directory) Initial(ctx context.Context) ([]domain.Station, error) {
	return d.base(ctx)
}

// Search resolves in priority order:
//  1. corpus loaded  -> local fuzzy match (typo-tolerant, no network)
//  2. no corpus      -> online per-query search (today's behaviour)
//  3. online errors  -> bundled starter list as an offline floor
func (d *Directory) Search(ctx context.Context, query string) ([]domain.Station, error) {
	if s := d.snapshot(); len(s) > 0 {
		return matchLocal(query, s), nil
	}
	if d.Online != nil {
		if st, err := d.Online.Search(ctx, query); err == nil {
			return st, nil
		}
	}
	if d.Offline != nil {
		st, err := d.Offline.Search(ctx, "")
		if err != nil {
			return nil, err
		}
		return matchLocal(query, st), nil
	}
	return nil, nil
}

// Popular returns the corpus sorted by community votes (local; no network).
func (d *Directory) Popular(ctx context.Context) ([]domain.Station, error) {
	stations, err := d.base(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Station, len(stations))
	copy(out, stations)
	sort.SliceStable(out, func(a, b int) bool { return out[a].Votes > out[b].Votes })
	if len(out) > 100 {
		out = out[:100]
	}
	return out, nil
}

// Refresh downloads a fresh full dump, replaces the corpus, and persists it.
// Returns the new corpus. The only method that uses the network.
func (d *Directory) Refresh(ctx context.Context) ([]domain.Station, error) {
	f, ok := d.Online.(fullFetcher)
	if !ok || d.Online == nil {
		return d.snapshot(), errors.New("online source cannot fetch the full dump")
	}
	stations, err := f.FetchAll(ctx)
	if err != nil {
		return d.snapshot(), err
	}
	d.setCorpus(stations)
	if d.Corpus != nil {
		_ = d.Corpus.Save(stations)
	}
	return stations, nil
}
