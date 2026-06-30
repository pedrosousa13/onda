package update

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckHitsNetworkThenCaches(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Write([]byte(`{"tag_name":"v9.9.9","assets":[
			{"name":"onda_9.9.9_linux_amd64.tar.gz","browser_download_url":"AURL"},
			{"name":"checksums.txt","browser_download_url":"SURL"}]}`))
	}))
	defer srv.Close()

	oldBase, oldNow, oldGOOS, oldGOARCH := apiBase, nowFn, goosFn, goarchFn
	defer func() { apiBase, nowFn, goosFn, goarchFn = oldBase, oldNow, oldGOOS, oldGOARCH }()
	apiBase = srv.URL
	goosFn = func() string { return "linux" }
	goarchFn = func() string { return "amd64" }
	nowFn = func() time.Time { return time.Unix(1000, 0) }

	dir := t.TempDir()
	st, err := Check(context.Background(), "1.0.0", dir)
	if err != nil {
		t.Fatal(err)
	}
	if !st.Available || st.Latest != "v9.9.9" || st.AssetURL != "AURL" {
		t.Fatalf("bad status: %+v", st)
	}

	// Second call within TTL must not hit the network.
	if _, err := Check(context.Background(), "1.0.0", dir); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 network call, got %d", calls)
	}
}

func TestCheckSilentOnHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()
	oldBase, oldNow := apiBase, nowFn
	defer func() { apiBase, nowFn = oldBase, oldNow }()
	apiBase = srv.URL
	nowFn = func() time.Time { return time.Unix(1000, 0) }

	st, err := Check(context.Background(), "1.0.0", t.TempDir())
	if err != nil {
		t.Fatalf("rate-limit/403 must be silent, got err: %v", err)
	}
	if st.Available {
		t.Fatal("no update should be reported on failed check")
	}
}
