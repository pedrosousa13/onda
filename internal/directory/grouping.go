package directory

import (
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

func key(r record) string {
	return strings.ToLower(strings.TrimSpace(r.Name)) + "|" +
		strings.ToLower(strings.TrimSpace(r.Homepage))
}

// GroupRecords merges single-stream records into logical stations with variants.
func GroupRecords(recs []record) []domain.Station {
	idx := map[string]int{}
	out := []domain.Station{}
	for _, r := range recs {
		if r.URL == "" {
			continue
		}
		k := key(r)
		i, ok := idx[k]
		if !ok {
			idx[k] = len(out)
			out = append(out, domain.Station{
				Name: r.Name, Country: r.Country, Tags: r.Tags,
				Homepage: r.Homepage, Lat: r.Lat, Lon: r.Lon,
			})
			i = idx[k]
		}
		out[i].Variants = append(out[i].Variants, domain.StreamVariant{
			URL: r.URL, Codec: r.Codec, Bitrate: r.Bitrate, HLS: r.HLS,
		})
	}
	sort.SliceStable(out, func(a, b int) bool { return out[a].Name < out[b].Name })
	return out
}
