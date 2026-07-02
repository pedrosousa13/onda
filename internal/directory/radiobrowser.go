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

const (
	// searchTimeout bounds a per-query search: small payload, must feel snappy.
	searchTimeout = 8 * time.Second
	// dumpStallTimeout aborts a full-catalog download only if it goes silent for
	// this long. The dump is ~70MB and streams for far longer than searchTimeout
	// on a normal link, so it must NOT share the search client's absolute cap —
	// that cap is what stranded users on the tiny embedded list. Progress, not a
	// whole-request clock, is what proves the mirror is alive.
	dumpStallTimeout = 30 * time.Second
	// fullDumpPath is the full-catalogue query. The explicit high limit is
	// required: /json/stations defaults to only 1000 rows.
	fullDumpPath = "/json/stations?hidebroken=true&limit=100000"
)

type RadioBrowser struct {
	mirrors    []string
	ua         string
	client     *http.Client // short absolute timeout: per-query search
	bulkClient *http.Client // no absolute timeout: the multi-tens-of-MB dump
	stall      time.Duration
}

func NewRadioBrowser(o RBOptions) *RadioBrowser {
	client, bulk := o.Client, o.Client
	if client == nil {
		client = &http.Client{Timeout: searchTimeout}
		// Clone the default transport (keeps proxy-from-env, dial timeouts, and
		// connection pooling) and add only a header timeout, so a dead mirror is
		// skipped fast without capping the body transfer of a healthy one.
		tr := http.DefaultTransport.(*http.Transport).Clone()
		tr.ResponseHeaderTimeout = 15 * time.Second
		bulk = &http.Client{Transport: tr}
	}
	return &RadioBrowser{mirrors: o.Mirrors, ua: o.UserAgent, client: client, bulkClient: bulk, stall: dumpStallTimeout}
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
// response is ~70MB of uncompressed JSON, so it uses the long-lived bulkClient
// and streams the decode rather than buffering the whole body.
func (rb *RadioBrowser) FetchAll(ctx context.Context) ([]domain.Station, error) {
	raw, err := rb.fetchDump(ctx, fullDumpPath, nil)
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

// fetchDump streams the full catalogue from the first working mirror, decoding
// as bytes arrive so the whole ~70MB is never held in memory at once. Mirrors
// are tried one at a time (not raced) so only the winning mirror drives
// onProgress and byte counts stay monotonic; a mirror that goes silent for
// longer than rb.stall is abandoned so a stuck socket can't hang forever.
func (rb *RadioBrowser) fetchDump(ctx context.Context, path string, onProgress func(int64)) ([]rbStation, error) {
	if len(rb.mirrors) == 0 {
		return nil, errors.New("no mirrors configured")
	}
	var lastErr error
	for _, base := range rb.mirrors {
		raw, err := rb.fetchDumpFrom(ctx, base+path, onProgress)
		if err != nil {
			lastErr = err
			continue // next mirror restarts its own byte counter from zero
		}
		return raw, nil
	}
	return nil, lastErr
}

func (rb *RadioBrowser) fetchDumpFrom(ctx context.Context, url string, onProgress func(int64)) ([]rbStation, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", rb.ua)
	resp, err := rb.bulkClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mirror %s: status %d", url, resp.StatusCode)
	}

	// Inactivity watchdog: cancel the request if no bytes arrive for rb.stall,
	// resetting the timer on every read that carries data. This bounds a stalled
	// socket without imposing a whole-request deadline on a slow-but-working link.
	watchdog := time.AfterFunc(rb.stall, cancel)
	defer watchdog.Stop()
	cr := &countingReader{r: resp.Body, onN: func(n int64) {
		watchdog.Reset(rb.stall)
		if onProgress != nil {
			onProgress(n)
		}
	}}

	var raw []rbStation
	if err := json.NewDecoder(cr).Decode(&raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// FetchAllWithProgress downloads the entire station list like FetchAll, but
// reports cumulative bytes downloaded via onProgress as they arrive.
func (rb *RadioBrowser) FetchAllWithProgress(ctx context.Context, onProgress func(int64)) ([]domain.Station, error) {
	raw, err := rb.fetchDump(ctx, fullDumpPath, onProgress)
	if err != nil {
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
