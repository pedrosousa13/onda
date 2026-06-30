//go:build !windows

package player

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"time"
)

func ipcAddress() string {
	return filepath.Join(os.TempDir(), "onda-mpv.sock")
}

func cleanupIPC(addr string) { _ = os.Remove(addr) }

func dialWithRetry(addr string) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultDialTimeout)
	defer cancel()
	var d net.Dialer
	for {
		conn, err := d.DialContext(ctx, "unix", addr)
		if err == nil {
			return conn, nil
		}
		select {
		case <-ctx.Done():
			return nil, errors.New("timed out connecting to mpv IPC socket")
		case <-time.After(50 * time.Millisecond): // brief backoff, matches the Windows path
		}
	}
}
