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
	Language    string  `json:"language"`
	ClickTrend  int     `json:"clicktrend"`
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

// FetchAll downloads the entire station list to build the local corpus. The
// HTTP transport negotiates gzip transparently, so the transfer is compressed.
// The explicit high limit is required: /json/stations defaults to only 1000
// rows, which would leave the "full" corpus a tiny fraction of the catalogue.
func (rb *RadioBrowser) FetchAll(ctx context.Context) ([]domain.Station, error) {
	raw, err := rb.fetchRaw(ctx, "/json/stations?hidebroken=true&limit=100000")
	if err != nil {
		return nil, err
	}
	return GroupRecords(rbToRecords(raw)), nil
}

// countingReader wraps a reader and reports the cumulative byte count read so
// far to onN after every non-empty Read.
type countingReader struct {
	r   io.Reader
	n   int64
	onN func(int64)
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	if n > 0 {
		c.n += int64(n)
		if c.onN != nil {
			c.onN(c.n)
		}
	}
	return n, err
}

// getSequentialProgress tries mirrors one at a time, in order, unlike
// getWithFallback which races them concurrently. Racing would let multiple
// in-flight bodies report progress simultaneously, producing interleaved,
// non-monotonic byte counts. Trying mirrors sequentially means only the
// winning mirror ever drives onProgress, so counts stay monotonic.
func (rb *RadioBrowser) getSequentialProgress(ctx context.Context, path string, onProgress func(int64)) ([]byte, error) {
	if len(rb.mirrors) == 0 {
		return nil, errors.New("no mirrors configured")
	}
	var lastErr error
	for _, base := range rb.mirrors {
		body, err := func() ([]byte, error) {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+path, nil)
			if err != nil {
				return nil, err
			}
			req.Header.Set("User-Agent", rb.ua)
			resp, err := rb.client.Do(req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("mirror %s: status %d", base, resp.StatusCode)
			}
			cr := &countingReader{r: resp.Body, onN: onProgress}
			return io.ReadAll(cr)
		}()
		if err != nil {
			lastErr = err
			continue // reset: next mirror starts its own counter from zero
		}
		return body, nil
	}
	return nil, lastErr
}

// FetchAllWithProgress downloads the entire station list like FetchAll, but
// reports cumulative decompressed bytes downloaded via onProgress as they
// arrive. It uses a sequential single-mirror path (see getSequentialProgress)
// so progress reporting stays monotonic.
func (rb *RadioBrowser) FetchAllWithProgress(ctx context.Context, onProgress func(int64)) ([]domain.Station, error) {
	body, err := rb.getSequentialProgress(ctx, "/json/stations?hidebroken=true&limit=100000", onProgress)
	if err != nil {
		return nil, err
	}
	var raw []rbStation
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	return GroupRecords(rbToRecords(raw)), nil
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
			Votes: s.Votes, ClickCount: s.ClickCount, Language: s.Language, Trend: s.ClickTrend,
		})
	}
	return recs
}
