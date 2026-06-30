package update

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestVerifyChecksum(t *testing.T) {
	data := []byte("hello onda")
	sum := sha256.Sum256(data)
	line := hex.EncodeToString(sum[:]) + "  onda_1.0.0_linux_amd64.tar.gz\n" +
		"deadbeef  other_file\n"

	if err := verifyChecksum(data, "onda_1.0.0_linux_amd64.tar.gz", line); err != nil {
		t.Fatalf("valid checksum rejected: %v", err)
	}
	if err := verifyChecksum([]byte("tampered"), "onda_1.0.0_linux_amd64.tar.gz", line); err == nil {
		t.Fatal("tampered data accepted")
	}
	if err := verifyChecksum(data, "missing.tar.gz", line); err == nil {
		t.Fatal("missing filename should error")
	}
}
