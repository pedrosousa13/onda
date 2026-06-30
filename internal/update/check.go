package update

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"runtime"
	"time"
)

// Injected for tests.
var (
	apiBase  = "https://api.github.com"
	nowFn    = time.Now
	goosFn   = func() string { return runtime.GOOS }
	goarchFn = func() string { return runtime.GOARCH }
)

const repoPath = "/repos/pedrosousa13/onda/releases/latest"

// Status is the result of an update check.
type Status struct {
	Current       string
	Latest        string
	Available     bool
	SelfUpdatable bool
	InstallKind   string
	AssetURL      string
	ChecksumsURL  string
}

type ghRelease struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// Check returns update status using a ≤24h cache. Network call only on cache miss.
// A failed/offline/rate-limited check is silent (returns a non-Available Status, nil error).
func Check(ctx context.Context, current, cacheDir string) (Status, error) {
	kind, writable := installKind()
	st := Status{Current: current, InstallKind: kind}

	c, ok := loadCache(cacheDir)
	if !ok || !fresh(c, nowFn()) {
		if fetched, err := fetchLatest(ctx, current); err == nil {
			c = fetched
			c.CheckedAt = nowFn()
			_ = saveCache(cacheDir, c)
		} else if !ok {
			return st, nil // first run, network failed: silent
		}
		// else: keep stale cache rather than nothing
	}

	st.Latest = c.LatestVersion
	st.AssetURL = c.AssetURL
	st.ChecksumsURL = c.ChecksumsURL
	st.Available = isNewer(current, c.LatestVersion)
	st.SelfUpdatable = st.Available && kind == "binary" && writable && st.AssetURL != ""
	return st, nil
}

func fetchLatest(ctx context.Context, current string) (cache, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiBase+repoPath, nil)
	if err != nil {
		return cache{}, err
	}
	req.Header.Set("User-Agent", "onda/"+current)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return cache{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return cache{}, &httpError{resp.StatusCode}
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return cache{}, err
	}
	var rel ghRelease
	if err := json.Unmarshal(body, &rel); err != nil {
		return cache{}, err
	}
	return cache{
		LatestVersion: rel.TagName,
		AssetURL:      selectAsset(rel.Assets, goosFn(), goarchFn()),
		ChecksumsURL:  checksumsURL(rel.Assets),
	}, nil
}

type httpError struct{ code int }

func (e *httpError) Error() string { return http.StatusText(e.code) }
