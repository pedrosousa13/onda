package directory

import (
	"sort"
	"strings"

	"github.com/pedrosousa13/radio/internal/domain"
)

// rankByQuery reorders stations best-match-first for the query, scoring name
// highest, then country, then tags. Stable, so equal scores keep their order.
func rankByQuery(query string, stations []domain.Station) []domain.Station {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return stations
	}
	type scored struct {
		s domain.Station
		n int
	}
	arr := make([]scored, len(stations))
	for i, s := range stations {
		arr[i] = scored{s, stationScore(q, s)}
	}
	sort.SliceStable(arr, func(a, b int) bool { return arr[a].n > arr[b].n })
	out := make([]domain.Station, len(arr))
	for i, x := range arr {
		out[i] = x.s
	}
	return out
}

func stationScore(q string, s domain.Station) int {
	best := fieldScore(q, s.Name) * 3 // name weighted highest
	if c := fieldScore(q, s.Country) * 2; c > best {
		best = c
	}
	for _, t := range s.Tags {
		if ts := fieldScore(q, t); ts > best {
			best = ts
		}
	}
	return best
}

func fieldScore(q, field string) int {
	f := strings.ToLower(field)
	switch {
	case f == q:
		return 100
	case strings.HasPrefix(f, q):
		return 80
	case strings.Contains(f, q):
		return 60
	case isSubsequence(q, f):
		return 30 // typo / out-of-order tolerance
	default:
		return 0
	}
}

// isSubsequence reports whether q's runes appear in order within f.
func isSubsequence(q, f string) bool {
	i := 0
	qr := []rune(q)
	for _, r := range f {
		if i < len(qr) && r == qr[i] {
			i++
		}
	}
	return i == len(qr)
}
