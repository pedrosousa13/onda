package directory

import (
	"sort"
	"strings"

	"github.com/pedrosousa13/onda/internal/domain"
)

// facetCounts groups stations by a key function, returns count-desc facets (tie: name asc).
func facetCounts(stations []domain.Station, key func(domain.Station) (name string, ok bool)) []domain.Facet {
	counts := map[string]int{}
	for _, s := range stations {
		if name, ok := key(s); ok && name != "" {
			counts[name]++
		}
	}
	out := make([]domain.Facet, 0, len(counts))
	for name, c := range counts {
		out = append(out, domain.Facet{Name: name, Count: c})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func countryFacets(sts []domain.Station) []domain.Facet {
	return facetCounts(sts, func(s domain.Station) (string, bool) { return s.Country, s.Country != "" })
}

func languageFacets(sts []domain.Station) []domain.Facet {
	return facetCounts(sts, func(s domain.Station) (string, bool) { return s.Language, s.Language != "" })
}

// tagFacets counts each tag occurrence across stations, capped to the top 100.
func tagFacets(sts []domain.Station) []domain.Facet {
	counts := map[string]int{}
	for _, s := range sts {
		for _, t := range s.Tags {
			t = strings.TrimSpace(t)
			if t != "" {
				counts[t]++
			}
		}
	}
	out := make([]domain.Facet, 0, len(counts))
	for name, c := range counts {
		out = append(out, domain.Facet{Name: name, Count: c})
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Name < out[j].Name
	})
	if len(out) > 100 {
		out = out[:100]
	}
	return out
}

// axisMatches reports whether a station belongs to a facet value (case-insensitive).
func axisMatches(s domain.Station, axis domain.Axis, value string) bool {
	switch axis {
	case domain.AxisTag:
		for _, t := range s.Tags {
			if strings.EqualFold(strings.TrimSpace(t), value) {
				return true
			}
		}
		return false
	case domain.AxisLanguage:
		return strings.EqualFold(s.Language, value)
	default:
		return strings.EqualFold(s.Country, value)
	}
}

// sortStations orders stations by the chosen key/direction.
func sortStations(sts []domain.Station, s domain.Sort) {
	sort.SliceStable(sts, func(i, j int) bool {
		a, b := sts[i], sts[j]
		switch s.Key {
		case domain.SortName:
			if s.Descending() {
				return a.Name > b.Name
			}
			return a.Name < b.Name
		case domain.SortTrend:
			if s.Descending() {
				return a.Trend > b.Trend
			}
			return a.Trend < b.Trend
		default:
			if s.Descending() {
				return a.Votes > b.Votes
			}
			return a.Votes < b.Votes
		}
	})
}
