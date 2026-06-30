package update

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// verifyChecksum confirms sha256(data) matches the entry for filename in a
// checksums.txt body ("<hex>  <filename>" lines).
func verifyChecksum(data []byte, filename, checksums string) error {
	var want string
	for _, ln := range strings.Split(checksums, "\n") {
		f := strings.Fields(ln)
		if len(f) == 2 && f[1] == filename {
			want = strings.ToLower(f[0])
			break
		}
	}
	if want == "" {
		return fmt.Errorf("no checksum entry for %q", filename)
	}
	sum := sha256.Sum256(data)
	if got := hex.EncodeToString(sum[:]); got != want {
		return fmt.Errorf("checksum mismatch for %q: got %s want %s", filename, got, want)
	}
	return nil
}
