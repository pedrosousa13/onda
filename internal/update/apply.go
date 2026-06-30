package update

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"runtime"
	"strings"

	"github.com/minio/selfupdate"
)

// Apply downloads the matching release archive, verifies its SHA256 against
// checksums.txt, extracts the onda binary, and atomically replaces the running
// binary (rollback on failure; Windows running-exe rename handled by selfupdate).
func Apply(ctx context.Context, st Status) error {
	if !st.SelfUpdatable || st.AssetURL == "" || st.ChecksumsURL == "" {
		return errors.New("update: not self-updatable")
	}

	archive, err := download(ctx, st.AssetURL)
	if err != nil {
		return err
	}
	sums, err := download(ctx, st.ChecksumsURL)
	if err != nil {
		return err
	}

	assetName := nameFromURL(st.AssetURL)
	if err := verifyChecksum(archive, assetName, string(sums)); err != nil {
		return err
	}
	// TODO(signing): verify minisign/cosign signature here before trusting the archive.

	binName := "onda"
	if runtime.GOOS == "windows" {
		binName = "onda.exe"
	}
	bin, err := extractBinary(archive, assetName, binName)
	if err != nil {
		return err
	}
	return selfupdate.Apply(bytes.NewReader(bin), selfupdate.Options{})
}

func download(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, &httpError{resp.StatusCode}
	}
	return io.ReadAll(io.LimitReader(resp.Body, 100<<20))
}

func nameFromURL(url string) string {
	if i := strings.LastIndex(url, "/"); i >= 0 {
		return url[i+1:]
	}
	return url
}
