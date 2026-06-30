package directory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pedrosousa13/radio/internal/domain"
)

type RBOptions struct {
	Mirrors   []string // base URLs; tried in order
	UserAgent string
	Client    *http.Client
}

type RadioBrowser struct {
	mirrors []string
	ua      string
	client  *http.Client
}

func NewRadioBrowser(o RBOptions) *RadioBrowser {
	c := o.Client
	if c == nil {
		c = &http.Client{Timeout: 10 * time.Second}
	}
	return &RadioBrowser{mirrors: o.Mirrors, ua: o.UserAgent, client: c}
}

type rbStation struct {
	Name        string  `json:"name"`
	URLResolved string  `json:"url_resolved"`
	Homepage    string  `json:"homepage"`
	Codec       string  `json:"codec"`
	Bitrate     int     `json:"bitrate"`
	Country     string  `json:"country"`
	Tags        string  `json:"tags"`
	GeoLat      float64 `json:"geo_lat"`
	GeoLong     float64 `json:"geo_long"`
	HLS         int     `json:"hls"`
}

func (rb *RadioBrowser) Search(ctx context.Context, query string) ([]domain.Station, error) {
	path := "/json/stations/search?limit=200&hidebroken=true&name=" + url.QueryEscape(query)
	body, err := rb.getWithFallback(ctx, path)
	if err != nil {
		return nil, err
	}
	var raw []rbStation
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	return GroupRecords(rbToRecords(raw)), nil
}

func (rb *RadioBrowser) getWithFallback(ctx context.Context, path string) ([]byte, error) {
	var lastErr error
	for _, base := range rb.mirrors {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+path, nil)
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("User-Agent", rb.ua)
		resp, err := rb.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		b, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil || resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("mirror %s: status %d", base, resp.StatusCode)
			continue
		}
		return b, nil
	}
	if lastErr == nil {
		lastErr = errors.New("no mirrors configured")
	}
	return nil, lastErr
}

func rbToRecords(raw []rbStation) []record {
	recs := make([]record, 0, len(raw))
	for _, s := range raw {
		if s.URLResolved == "" {
			continue
		}
		var tags []string
		if s.Tags != "" {
			tags = strings.Split(s.Tags, ",")
		}
		recs = append(recs, record{
			Name: s.Name, Country: s.Country, Tags: tags, Homepage: s.Homepage,
			Lat: s.GeoLat, Lon: s.GeoLong, URL: s.URLResolved,
			Codec: s.Codec, Bitrate: s.Bitrate, HLS: s.HLS == 1,
		})
	}
	return recs
}
