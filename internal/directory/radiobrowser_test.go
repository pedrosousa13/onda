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
  {"name":"KEXP","homepage":"kexp.org","url_resolved":"u128","codec":"MP3","bitrate":128,"countrycode":"US","country":"United States","tags":"indie,seattle"},
  {"name":"KEXP","homepage":"kexp.org","url_resolved":"u64","codec":"MP3","bitrate":64,"country":"United States","tags":""}
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
	  {"name":"Radio Eins","url_resolved":"http://a","country":"Germany","bitrate":128,"votes":42},
	  {"name":"Jazz FM","url_resolved":"http://b","country":"United Kingdom","bitrate":64,"votes":3}
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
