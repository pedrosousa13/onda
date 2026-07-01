package directory

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

const rbSample = `[
  {"name":"KEXP","homepage":"kexp.org","url_resolved":"u128","codec":"MP3","bitrate":128,"countrycode":"US","country":"United States","tags":"indie,seattle","language":"English","clicktrend":3},
  {"name":"KEXP","homepage":"kexp.org","url_resolved":"u64","codec":"MP3","bitrate":64,"country":"United States","tags":"","language":"English","clicktrend":5}
]`

func TestRadioBrowserSearchGroups(t *testing.T) {
	var (
		mu    sync.Mutex
		gotUA string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		gotUA = r.Header.Get("User-Agent")
		mu.Unlock()
		_, _ = w.Write([]byte(rbSample))
	}))
	defer srv.Close()

	rb := NewRadioBrowser(RBOptions{Mirrors: []string{srv.URL}, UserAgent: "radio/test"})
	stations, err := rb.Search(context.Background(), "kexp")
	if err != nil {
		t.Fatal(err)
	}
	if len(stations) != 1 || len(stations[0].Variants) != 2 {
		t.Fatalf("want 1 station w/ 2 variants, got %+v", stations)
	}
	mu.Lock()
	ua := gotUA
	mu.Unlock()
	if ua != "radio/test" {
		t.Fatalf("missing/incorrect User-Agent: %q", ua)
	}
}

func TestRadioBrowserFallsOverMirrors(t *testing.T) {
	good := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(rbSample))
	}))
	defer good.Close()
	rb := NewRadioBrowser(RBOptions{
		Mirrors:   []string{"http://127.0.0.1:1", good.URL}, // first is dead
		UserAgent: "radio/test",
	})
	if _, err := rb.Search(context.Background(), "kexp"); err != nil {
		t.Fatalf("expected fallback to succeed, got %v", err)
	}
}

type dumpStubRT struct{ body string }

func (s dumpStubRT) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(s.body)),
		Header:     make(http.Header),
	}, nil
}

func TestFetchAllGroupsDump(t *testing.T) {
	const dump = `[
	  {"name":"Radio Eins","url_resolved":"http://a","country":"Germany","bitrate":128,"votes":42,"language":"German","clicktrend":2},
	  {"name":"Jazz FM","url_resolved":"http://b","country":"United Kingdom","bitrate":64,"votes":3,"language":"English","clicktrend":0}
	]`
	rb := NewRadioBrowser(RBOptions{
		Mirrors: []string{"http://example.test"},
		Client:  &http.Client{Transport: dumpStubRT{body: dump}},
	})
	out, err := rb.FetchAll(context.Background())
	if err != nil {
		t.Fatalf("FetchAll: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 stations, got %d", len(out))
	}
}

func TestCountingReaderReportsCumulativeBytes(t *testing.T) {
	var last int64
	cr := &countingReader{r: strings.NewReader("hello world"), onN: func(n int64) { last = n }}
	buf, err := io.ReadAll(cr)
	if err != nil {
		t.Fatal(err)
	}
	if int64(len(buf)) != last {
		t.Fatalf("reported %d, read %d", last, len(buf))
	}
	if last != 11 {
		t.Fatalf("want 11 bytes, got %d", last)
	}
}

func TestFetchAllRequestsHideBroken(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.RequestURI()
		w.Write([]byte("[]"))
	}))
	defer srv.Close()
	rb := NewRadioBrowser(RBOptions{Mirrors: []string{srv.URL}, UserAgent: "test"})
	_, _ = rb.FetchAll(context.Background())
	if !strings.Contains(gotPath, "hidebroken=true") {
		t.Fatalf("full-dump path missing hidebroken=true: %s", gotPath)
	}
	// A high limit is mandatory: without it /json/stations caps at 1000 rows,
	// which would make the "full" corpus a tiny fraction of the catalogue.
	if !strings.Contains(gotPath, "limit=100000") {
		t.Fatalf("full-dump path missing high limit: %s", gotPath)
	}
}

// chunkedBody yields the body one bounded chunk per Read, so the countingReader
// is driven across multiple Reads deterministically (unlike an httptest server,
// where the client transport may coalesce flushed writes into a single Read).
type chunkedBody struct {
	chunks [][]byte
	i      int
}

func (b *chunkedBody) Read(p []byte) (int, error) {
	if b.i >= len(b.chunks) {
		return 0, io.EOF
	}
	n := copy(p, b.chunks[b.i])
	b.chunks[b.i] = b.chunks[b.i][n:]
	if len(b.chunks[b.i]) == 0 {
		b.i++
	}
	return n, nil
}

func (b *chunkedBody) Close() error { return nil }

type chunkedRT struct{ chunks [][]byte }

func (rt chunkedRT) RoundTrip(*http.Request) (*http.Response, error) {
	cp := make([][]byte, len(rt.chunks))
	for i, c := range rt.chunks {
		cp[i] = append([]byte(nil), c...)
	}
	return &http.Response{
		StatusCode: 200,
		Body:       &chunkedBody{chunks: cp},
		Header:     make(http.Header),
	}, nil
}

func TestFetchAllWithProgressReportsMonotonicBytes(t *testing.T) {
	// A valid stations JSON array split into chunks. chunkedRT hands each chunk
	// out on a separate Read, so onProgress fires once per chunk and the
	// strictly-increasing assertion below is actually exercised.
	chunks := [][]byte{
		[]byte(`[{"name":"Radio Eins","url_resolved":"http://a","country":"Germany","bitrate":128,"votes":42,"language":"German","clicktrend":1},`),
		[]byte(`{"name":"Jazz FM","url_resolved":"http://b","country":"United Kingdom","bitrate":64,"votes":3,"language":"English","clicktrend":0},`),
		[]byte(`{"name":"BBC","url_resolved":"http://c","country":"United Kingdom","bitrate":320,"votes":7,"language":"English","clicktrend":2}]`),
	}
	var total int64
	for _, c := range chunks {
		total += int64(len(c))
	}

	rb := NewRadioBrowser(RBOptions{
		Mirrors:   []string{"http://example.test"},
		UserAgent: "test",
		Client:    &http.Client{Transport: chunkedRT{chunks: chunks}},
	})

	var reports []int64
	out, err := rb.FetchAllWithProgress(context.Background(), func(n int64) {
		reports = append(reports, n)
	})
	if err != nil {
		t.Fatalf("FetchAllWithProgress: %v", err)
	}
	if len(out) != 3 {
		t.Fatalf("expected 3 stations, got %d", len(out))
	}
	if len(reports) < 2 {
		t.Fatalf("expected multiple progress reports across chunks, got %d: %v", len(reports), reports)
	}
	for i := 1; i < len(reports); i++ {
		if reports[i] <= reports[i-1] {
			t.Fatalf("progress not strictly increasing: %v", reports)
		}
	}
	if reports[len(reports)-1] != total {
		t.Fatalf("final report %d != body length %d", reports[len(reports)-1], total)
	}
}

func TestFetchAllWithProgressFallsOverMirrors(t *testing.T) {
	good := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(rbSample))
	}))
	defer good.Close()
	rb := NewRadioBrowser(RBOptions{
		Mirrors:   []string{"http://127.0.0.1:1", good.URL}, // first is dead
		UserAgent: "radio/test",
	})
	if _, err := rb.FetchAllWithProgress(context.Background(), nil); err != nil {
		t.Fatalf("expected fallback to succeed, got %v", err)
	}
}
