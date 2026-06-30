package directory

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

const rbSample = `[
  {"name":"KEXP","homepage":"kexp.org","url_resolved":"u128","codec":"MP3","bitrate":128,"countrycode":"US","country":"United States","tags":"indie,seattle"},
  {"name":"KEXP","homepage":"kexp.org","url_resolved":"u64","codec":"MP3","bitrate":64,"country":"United States","tags":""}
]`

func TestRadioBrowserSearchGroups(t *testing.T) {
	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
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
	if gotUA != "radio/test" {
		t.Fatalf("missing/incorrect User-Agent: %q", gotUA)
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
