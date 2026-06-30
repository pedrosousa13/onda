//go:build integration

package player

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Requires a real mpv on PATH. Plays a tiny local file and expects a state event.
func TestPlayLocalFile(t *testing.T) {
	p, err := New(Options{})
	if err != nil {
		t.Skip("mpv not available:", err)
	}
	defer p.Close()

	fixture := filepath.Join("testdata", "silence.mp3")
	if _, err := os.Stat(fixture); err != nil {
		t.Skip("no fixture at", fixture)
	}
	if err := p.Play(fixture); err != nil {
		t.Fatal(err)
	}
	select {
	case e := <-p.Events():
		_ = e
	case <-time.After(5 * time.Second):
		t.Fatal("no playback event within 5s")
	}
}
