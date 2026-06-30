//go:build windows

package player

import (
	"context"
	"net"
	"time"

	"github.com/Microsoft/go-winio"
)

func ipcAddress() string { return `\\.\pipe\onda-mpv` }

func cleanupIPC(string) {} // named pipes vanish when mpv exits; nothing to remove

func dialWithRetry(addr string) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultDialTimeout)
	defer cancel()
	for {
		conn, err := winio.DialPipeContext(ctx, addr)
		if err == nil {
			return conn, nil
		}
		select {
		case <-ctx.Done():
			return nil, err
		case <-time.After(50 * time.Millisecond):
		}
	}
}
