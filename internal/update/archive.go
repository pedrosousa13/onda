package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"
)

// extractBinary returns the bytes of binName from a release archive. assetName
// selects the format (.zip vs .tar.gz).
func extractBinary(archive []byte, assetName, binName string) ([]byte, error) {
	if strings.HasSuffix(assetName, ".zip") {
		return extractZip(archive, binName)
	}
	return extractTarGz(archive, binName)
}

func extractTarGz(archive []byte, binName string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if h.Typeflag == tar.TypeReg && baseName(h.Name) == binName {
			return io.ReadAll(tr)
		}
	}
	return nil, fmt.Errorf("%q not found in archive", binName)
}

func extractZip(archive []byte, binName string) ([]byte, error) {
	zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
	if err != nil {
		return nil, err
	}
	for _, f := range zr.File {
		if baseName(f.Name) == binName {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
	}
	return nil, fmt.Errorf("%q not found in archive", binName)
}

func baseName(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[i+1:]
	}
	return p
}
