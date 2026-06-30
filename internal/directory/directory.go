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
