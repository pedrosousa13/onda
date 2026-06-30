package directory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pedrosousa13/onda/internal/domain"
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
		c = &http.Client{Timeout: 8 * time.Second}
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
	Votes       int     `json:"votes"`
	ClickCount  int     `json:"clickcount"`
}

func (rb *RadioBrowser) fetchRaw(ctx context.Context, path string) ([]rbStation, error) {
	body, err := rb.getWithFallback(ctx, path)
	if err != nil {
		return nil, err
	}
	var raw []rbStation
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// Search matches the query against name, tag, and country in parallel, merges
// and de-duplicates the results, then ranks them best-match-first (fuzzy).
func (rb *RadioBrowser) Search(ctx context.Context, query string) ([]domain.Station, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		raw, err := rb.fetchRaw(ctx, "/json/stations/search?limit=200&hidebroken=true")
		if err != nil {
			return nil, err
		}
		return GroupRecords(rbToRecords(raw)), nil
	}

	enc := url.QueryEscape(q)
	paths := []string{
		"/json/stations/search?limit=80&hidebroken=true&name=" + enc,
		"/json/stations/search?limit=80&hidebroken=true&tag=" + enc,
		"/json/stations/search?limit=40&hidebroken=true&country=" + enc,
	}
	type res struct {
		raw []rbStation
		err error
	}
	ch := make(chan res, len(paths))
	for _, p := range paths {
		go func(p string) {
			raw, err := rb.fetchRaw(ctx, p)
			ch <- res{raw: raw, err: err}
		}(p)
	}
	var all []rbStation
	var anyOK bool
	var lastErr error
	for range paths {
		r := <-ch
		if r.err != nil {
			lastErr = r.err
			continue
		}
		anyOK = true
		all = append(all, r.raw...)
	}
	if !anyOK {
		return nil, lastErr
	}
	return rankByQuery(q, GroupRecords(rbToRecords(all))), nil
}

// TopVoted returns the most up-voted stations (community popularity). This is a
// read-only GET — it reports nothing about the user, unlike the /vote endpoint.
func (rb *RadioBrowser) TopVoted(ctx context.Context, limit int) ([]domain.Station, error) {
	path := "/json/stations/topvote?hidebroken=true&limit=" + strconv.Itoa(limit)
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

// getWithFallback races all mirrors concurrently and returns the first success,
// cancelling the rest. Effective latency is the fastest mirror, and a dead or
// slow mirror no longer blocks the others.
func (rb *RadioBrowser) getWithFallback(ctx context.Context, path string) ([]byte, error) {
	if len(rb.mirrors) == 0 {
		return nil, errors.New("no mirrors configured")
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type result struct {
		body []byte
		err  error
	}
	ch := make(chan result, len(rb.mirrors))

	for _, base := range rb.mirrors {
		go func(base string) {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+path, nil)
			if err != nil {
				ch <- result{err: err}
				return
			}
			req.Header.Set("User-Agent", rb.ua)
			resp, err := rb.client.Do(req)
			if err != nil {
				ch <- result{err: err}
				return
			}
			b, readErr := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if readErr != nil {
				ch <- result{err: readErr}
				return
			}
			if resp.StatusCode != http.StatusOK {
				ch <- result{err: fmt.Errorf("mirror %s: status %d", base, resp.StatusCode)}
				return
			}
			ch <- result{body: b}
		}(base)
	}

	var lastErr error
	for range rb.mirrors {
		r := <-ch
		if r.err == nil {
			return r.body, nil // first success wins; defer cancel() stops the rest
		}
		lastErr = r.err
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
			Votes: s.Votes, ClickCount: s.ClickCount,
		})
	}
	return recs
}
