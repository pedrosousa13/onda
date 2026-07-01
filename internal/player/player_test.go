package player

import (
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestEncodeCommand(t *testing.T) {
	got, err := encodeCommand(7, "loadfile", "http://x/stream")
	if err != nil {
		t.Fatal(err)
	}
	want := `{"command":["loadfile","http://x/stream"],"request_id":7}` + "\n"
	if string(got) != want {
		t.Fatalf("want %q got %q", want, got)
	}
}

func TestParseLineEvent(t *testing.T) {
	f, err := parseLine([]byte(`{"event":"property-change","name":"media-title","data":"Now Playing"}`))
	if err != nil {
		t.Fatal(err)
	}
	if f.Event != "property-change" || f.Name != "media-title" || f.Data != "Now Playing" {
		t.Fatalf("unexpected frame: %+v", f)
	}
}

func TestParseLineIgnoresReplies(t *testing.T) {
	f, err := parseLine([]byte(`{"error":"success","request_id":7}`))
	if err != nil {
		t.Fatal(err)
	}
	if f.Event != "" {
		t.Fatalf("reply should have empty Event, got %q", f.Event)
	}
}

func TestParseLineEndFileReason(t *testing.T) {
	f, err := parseLine([]byte(`{"event":"end-file","reason":"error"}`))
	if err != nil {
		t.Fatal(err)
	}
	if f.Event != "end-file" || f.Reason != "error" {
		t.Fatalf("unexpected frame: %+v", f)
	}
}

func TestNewRequiresMpv(t *testing.T) {
	// With a bogus binary name, New must fail fast and not leak a process.
	_, err := New(Options{Binary: "definitely-not-mpv-xyz"})
	if err == nil {
		t.Fatal("expected error when mpv binary is missing")
	}
}

// TestNewDoesNotStartMpv guards the fix for onda grabbing the audio device at
// launch: New must only verify mpv is on PATH, not spawn it or touch IPC.
func TestNewDoesNotStartMpv(t *testing.T) {
	if _, err := lookPathMpv(); err != nil {
		t.Skip("mpv not available:", err)
	}
	p, err := New(Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	if p.started {
		t.Fatal("New must not start mpv")
	}
	if p.cmd != nil || p.conn != nil {
		t.Fatal("New must not spawn a process or open an IPC connection")
	}
	if _, err := os.Stat(p.sock); err == nil {
		t.Fatal("New must not create the IPC socket file")
	}

	// Events() must be usable before mpv starts (no panic on nil channel, no
	// spurious events since nothing has happened yet).
	select {
	case e := <-p.Events():
		t.Fatalf("unexpected event before Play: %+v", e)
	case <-time.After(100 * time.Millisecond):
	}
}

// TestPlayStartsMpv proves the other half of the lazy-start contract: calling
// Play() is what actually spawns mpv and connects IPC.
func TestPlayStartsMpv(t *testing.T) {
	if _, err := lookPathMpv(); err != nil {
		t.Skip("mpv not available:", err)
	}
	p, err := New(Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	if err := p.Play("http://example.invalid/stream"); err != nil {
		t.Fatal(err)
	}

	p.mu.Lock()
	started := p.started
	hasConn := p.conn != nil
	p.mu.Unlock()
	if !started || !hasConn {
		t.Fatal("Play must start mpv and connect IPC")
	}
	if _, err := os.Stat(p.sock); err != nil {
		t.Fatalf("expected IPC socket to exist after Play: %v", err)
	}
}

// TestNoopsBeforeStart ensures control methods degrade to no-ops instead of
// erroring, blocking, or starting mpv when called before Play has ever run.
func TestNoopsBeforeStart(t *testing.T) {
	if _, err := lookPathMpv(); err != nil {
		t.Skip("mpv not available:", err)
	}
	p, err := New(Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	if err := p.Stop(); err != nil {
		t.Fatalf("Stop before start: %v", err)
	}
	if err := p.Pause(); err != nil {
		t.Fatalf("Pause before start: %v", err)
	}
	if err := p.Resume(); err != nil {
		t.Fatalf("Resume before start: %v", err)
	}
	if err := p.Close(); err != nil {
		t.Fatalf("Close before start: %v", err)
	}
	if p.started {
		t.Fatal("no-op control methods must not start mpv")
	}
}

// TestVolumeBeforeStartIsStored verifies Volume() stores the desired volume
// instead of erroring or starting mpv when called before Play.
func TestVolumeBeforeStartIsStored(t *testing.T) {
	if _, err := lookPathMpv(); err != nil {
		t.Skip("mpv not available:", err)
	}
	p, err := New(Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	if err := p.Volume(42); err != nil {
		t.Fatal(err)
	}
	if p.started {
		t.Fatal("Volume must not start mpv")
	}
	if p.volume == nil || *p.volume != 42 {
		t.Fatalf("expected stored volume 42, got %v", p.volume)
	}
}

func lookPathMpv() (string, error) {
	return exec.LookPath("mpv")
}
