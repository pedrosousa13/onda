package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"testing"
)

func makeTarGz(t *testing.T, name string, body []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: name, Size: int64(len(body)), Mode: 0o755}); err != nil {
		t.Fatal(err)
	}
	tw.Write(body)
	tw.Close()
	gz.Close()
	return buf.Bytes()
}

func makeZip(t *testing.T, name string, body []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	w, _ := zw.Create(name)
	w.Write(body)
	zw.Close()
	return buf.Bytes()
}

func TestExtractBinary(t *testing.T) {
	body := []byte("BINARY")

	tgz := makeTarGz(t, "onda", body)
	got, err := extractBinary(tgz, "onda_1.0.0_linux_amd64.tar.gz", "onda")
	if err != nil || !bytes.Equal(got, body) {
		t.Fatalf("tar.gz extract: %v / %q", err, got)
	}

	z := makeZip(t, "onda.exe", body)
	got, err = extractBinary(z, "onda_1.0.0_windows_amd64.zip", "onda.exe")
	if err != nil || !bytes.Equal(got, body) {
		t.Fatalf("zip extract: %v / %q", err, got)
	}

	if _, err := extractBinary(tgz, "onda_1.0.0_linux_amd64.tar.gz", "missing"); err == nil {
		t.Fatal("missing binary should error")
	}
}
