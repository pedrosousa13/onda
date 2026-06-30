package directory

import (
	"sort"
	"strings"

	"github.com/pedrosousa13/onda/internal/domain"
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
	case fuzzyTokenMatch(q, f):
		return 20 // typo tolerance (edit distance)
	default:
		return 0
	}
}

// fuzzyTokenMatch reports whether q is within Damerau-Levenshtein distance 2 of
// any whitespace-separated token in f. Gated to queries of at least 4 runes so
// short queries — already served by the contains/subsequence tiers — don't match
// unrelated noise. f is assumed already lower-cased.
func fuzzyTokenMatch(q, f string) bool {
	qr := []rune(q)
	if len(qr) < 4 {
		return false
	}
	for _, tok := range strings.Fields(f) {
		tr := []rune(tok)
		// The distance can't be <=2 if the lengths differ by more than 2.
		if d := len(tr) - len(qr); d > 2 || d < -2 {
			continue
		}
		if damerauLevenshtein(qr, tr) <= 2 {
			return true
		}
	}
	return false
}

// damerauLevenshtein returns the optimal string alignment distance between a and
// b — Levenshtein edits plus adjacent transpositions — which captures the common
// typo of two swapped letters (e.g. "raido" → "radio") at distance 1.
func damerauLevenshtein(a, b []rune) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	prev2 := make([]int, lb+1) // row i-2
	prev := make([]int, lb+1)  // row i-1
	curr := make([]int, lb+1)  // row i
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			m := prev[j] + 1 // deletion
			if ins := curr[j-1] + 1; ins < m {
				m = ins
			}
			if sub := prev[j-1] + cost; sub < m {
				m = sub
			}
			if i > 1 && j > 1 && a[i-1] == b[j-2] && a[i-2] == b[j-1] {
				if t := prev2[j-2] + 1; t < m { // transposition
					m = t
				}
			}
			curr[j] = m
		}
		prev2, prev, curr = prev, curr, prev2
	}
	return prev[lb]
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
