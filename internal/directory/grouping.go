package directory

import (
	"regexp"
	"sort"
	"strings"

	"github.com/pedrosousa13/radio/internal/domain"
)

// record is a normalized single-stream entry from any source.
type record struct {
	Name     string
	Country  string
	Tags     []string
	Homepage string
	Lat, Lon float64
	URL      string
	Codec    string
	Bitrate  int
	HLS      bool
}

var (
	bracketRe  = regexp.MustCompile(`[(\[][^)\]]*[)\]]`) // (Hi-Fi), (hifi.aac), (metadata), [aac], [mp3]
	nonAlphaRe = regexp.MustCompile(`[^a-z0-9]+`)        // punctuation/spacing → single space
	hifiRe     = regexp.MustCompile(`(?i)hi[-.\s]?fi|lossless|flac`)
)

// formatTokens are codec/format words dropped from a normalized name so that
// e.g. "FIP" and "FIP aac" merge.
var formatTokens = map[string]bool{
	"aac": true, "mp3": true, "flac": true, "hifi": true, "ogg": true, "hd": true,
}

// broadcasterPrefixes are public-broadcaster acronyms stripped from the START of
// a name so "RTP Antena 3" ≈ "Antena 3". Conservative list; country stays in the
// key, so cross-country collisions can't happen.
var broadcasterPrefixes = map[string]bool{
	"rtp": true, "bbc": true, "npr": true, "rai": true, "rte": true, "rtve": true,
	"ard": true, "zdf": true, "orf": true, "srf": true, "yle": true, "nrk": true,
}

// trailingNoise are suffix words dropped from a name ("Antena 3 - Main" → "Antena 3").
var trailingNoise = map[string]bool{
	"main": true, "hd": true, "stream": true, "official": true,
}

// groupKey merges variants of the same station: it strips quality/format
// parentheticals and punctuation from the name, then keys on name + country.
// So "FIP Jazz", "FIP Jazz (Hi-Fi)" and "FIP Jazz (hifi.aac)" collapse to one.
func groupKey(r record) string {
	return normalizeName(r.Name) + "|" + strings.ToLower(strings.TrimSpace(r.Country))
}

func normalizeName(name string) string {
	s := strings.ToLower(name)
	s = bracketRe.ReplaceAllString(s, " ")
	s = nonAlphaRe.ReplaceAllString(s, " ")
	words := strings.Fields(s)
	kept := words[:0]
	for _, w := range words {
		if formatTokens[w] {
			continue
		}
		kept = append(kept, w)
	}
	// Strip a leading broadcaster acronym ("RTP Antena 3" → "Antena 3").
	if len(kept) > 1 && broadcasterPrefixes[kept[0]] {
		kept = kept[1:]
	}
	// Strip trailing noise words ("Antena 3 Main" → "Antena 3").
	for len(kept) > 1 && trailingNoise[kept[len(kept)-1]] {
		kept = kept[:len(kept)-1]
	}
	out := strings.Join(kept, " ")
	if out == "" { // name was only format tokens — fall back to punctuation-stripped form
		return strings.TrimSpace(s)
	}
	return out
}

// displayName strips quality parentheticals/brackets but keeps original casing.
func displayName(name string) string {
	s := bracketRe.ReplaceAllString(name, "")
	return strings.Join(strings.Fields(s), " ")
}

func isHiFi(r record) bool {
	return hifiRe.MatchString(r.Name) || strings.Contains(strings.ToLower(r.Codec), "flac")
}

// GroupRecords merges single-stream records into logical stations with variants,
// with each station's variants sorted best-quality-first.
func GroupRecords(recs []record) []domain.Station {
	idx := map[string]int{}
	out := []domain.Station{}
	for _, r := range recs {
		if r.URL == "" {
			continue
		}
		k := groupKey(r)
		i, ok := idx[k]
		if !ok {
			idx[k] = len(out)
			out = append(out, domain.Station{
				Name: displayName(r.Name), Country: r.Country, Tags: r.Tags,
				Homepage: r.Homepage, Lat: r.Lat, Lon: r.Lon,
			})
			i = idx[k]
		}
		out[i].Variants = append(out[i].Variants, domain.StreamVariant{
			URL: r.URL, Codec: r.Codec, Bitrate: r.Bitrate, HLS: r.HLS, Lossless: isHiFi(r),
		})
	}
	for i := range out {
		sortVariants(out[i].Variants)
		out[i].Variants = dedupeVariants(out[i].Variants)
	}
	sort.SliceStable(out, func(a, b int) bool { return out[a].Name < out[b].Name })
	return out
}

// dedupeVariants keeps one variant per distinct quality label (e.g. collapses
// eight identical 192k mirror entries into a single "192k" choice). Assumes the
// slice is already sorted best-first, so the kept instance is the best one.
func dedupeVariants(vs []domain.StreamVariant) []domain.StreamVariant {
	seen := map[string]bool{}
	out := vs[:0]
	for _, v := range vs {
		q := v.Quality()
		if seen[q] {
			continue
		}
		seen[q] = true
		out = append(out, v)
	}
	return out
}

// sortVariants orders a station's variants best-quality-first (HiFi, then by bitrate desc).
func sortVariants(vs []domain.StreamVariant) {
	sort.SliceStable(vs, func(a, b int) bool {
		qa, qb := variantRank(vs[a]), variantRank(vs[b])
		return qa > qb
	})
}

func variantRank(v domain.StreamVariant) int {
	if v.Lossless {
		return 9999
	}
	return v.Bitrate
}
