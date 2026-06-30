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
	}
	sort.SliceStable(out, func(a, b int) bool { return out[a].Name < out[b].Name })
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
