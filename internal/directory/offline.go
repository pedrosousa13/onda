package directory

import (
	"context"
	_ "embed"
	"encoding/json"
	"strings"

	"github.com/pedrosousa13/radio/internal/domain"
)

//go:embed data/stations.json
var offlineJSON []byte

type rawStation struct {
	Name     string   `json:"name"`
	Country  string   `json:"country"`
	Tags     []string `json:"tags"`
	Homepage string   `json:"homepage"`
	URL      string   `json:"url"`
	Codec    string   `json:"codec"`
	Bitrate  int      `json:"bitrate"`
}

type Offline struct{ stations []domain.Station }

func NewOffline() *Offline {
	var raw []rawStation
	_ = json.Unmarshal(offlineJSON, &raw)
	return &Offline{stations: GroupRecords(toRecords(raw))}
}

func (o *Offline) Search(_ context.Context, q string) ([]domain.Station, error) {
	if q == "" {
		return o.stations, nil
	}
	q = strings.ToLower(q)
	var out []domain.Station
	for _, s := range o.stations {
		if strings.Contains(strings.ToLower(s.Name), q) ||
			strings.Contains(strings.ToLower(s.Country), q) {
			out = append(out, s)
		}
	}
	return out, nil
}

// --- TEMPORARY stub, replaced in Task 7 ---
func toRecords(raw []rawStation) []rawStation { return raw }

func GroupRecords(raw []rawStation) []domain.Station {
	out := make([]domain.Station, len(raw))
	for i, r := range raw {
		out[i] = domain.Station{
			Name: r.Name, Country: r.Country, Tags: r.Tags, Homepage: r.Homepage,
			Variants: []domain.StreamVariant{{URL: r.URL, Codec: r.Codec, Bitrate: r.Bitrate}},
		}
	}
	return out
}
