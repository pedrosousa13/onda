package player

import (
	"bufio"
	"errors"
	"net"
	"os/exec"
	"sync"
	"time"
)

const defaultDialTimeout = 3 * time.Second

// Event is emitted as playback state changes.
type Event struct {
	Kind  string // "title" | "playing" | "idle" | "error"
	Title string // set when Kind=="title"
	Err   error  // set when Kind=="error"
}

type Options struct {
	Binary    string // defaults to "mpv"
	Normalize bool   // start with loudness normalization (dynaudnorm) enabled
}

type Player struct {
	cmd    *exec.Cmd
	conn   net.Conn
	events chan Event
	mu     sync.Mutex
	id     int
	sock   string
}

// New verifies mpv exists, starts it headless, and connects to its IPC socket.
func New(opts Options) (*Player, error) {
	bin := opts.Binary
	if bin == "" {
		bin = "mpv"
	}
	if _, err := exec.LookPath(bin); err != nil {
		return nil, errors.New("mpv not found on PATH: install mpv (e.g. `brew install mpv`, or `scoop install mpv` on Windows)")
	}
	// ipcAddress() is platform-specific (Unix socket path or Windows pipe name).
	addr := ipcAddress()
	cleanupIPC(addr) // remove any stale socket file (no-op on Windows)
	args := []string{"--idle=yes", "--no-video", "--no-terminal", "--input-ipc-server=" + addr}
	if opts.Normalize {
		args = append(args, "--af=dynaudnorm")
	}
	cmd := exec.Command(bin, args...)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	conn, err := dialWithRetry(addr) // platform-specific dialer
	if err != nil {
		reap(cmd)
		return nil, err
	}
	p := &Player{cmd: cmd, conn: conn, events: make(chan Event, 16), sock: addr}
	if err := p.observe(); err != nil {
		_ = p.Close()
		return nil, err
	}
	go p.readLoop()
	return p, nil
}

func (p *Player) Events() <-chan Event { return p.events }

func (p *Player) Play(url string) error { return p.send("loadfile", url) }
func (p *Player) Stop() error           { return p.send("stop") }
func (p *Player) Pause() error          { return p.send("set_property", "pause", true) }
func (p *Player) Resume() error         { return p.send("set_property", "pause", false) }
func (p *Player) Volume(pct int) error  { return p.send("set_property", "volume", pct) }

// SetNormalize toggles loudness normalization live by setting mpv's audio-filter
// chain to dynaudnorm (on) or clearing it (off).
func (p *Player) SetNormalize(on bool) error {
	if on {
		return p.send("set_property", "af", "dynaudnorm")
	}
	return p.send("set_property", "af", "")
}

func (p *Player) observe() error {
	if err := p.send("observe_property", 1, "media-title"); err != nil {
		return err
	}
	return p.send("observe_property", 2, "core-idle")
}

func (p *Player) send(args ...any) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.id++
	b, err := encodeCommand(p.id, args...)
	if err != nil {
		return err
	}
	_, err = p.conn.Write(b)
	return err
}

func (p *Player) readLoop() {
	sc := bufio.NewScanner(p.conn)
	for sc.Scan() {
		f, err := parseLine(sc.Bytes())
		if err != nil || f.Event == "" {
			continue
		}
		switch {
		case f.Event == "property-change" && f.Name == "media-title":
			if title, ok := f.Data.(string); ok {
				p.emit(Event{Kind: "title", Title: title})
			}
		case f.Event == "property-change" && f.Name == "core-idle":
			if idle, ok := f.Data.(bool); ok && idle {
				p.emit(Event{Kind: "idle"})
			} else {
				p.emit(Event{Kind: "playing"})
			}
		case f.Event == "end-file":
			if f.Reason == "error" {
				p.emit(Event{Kind: "error", Err: errors.New("stream failed to load")})
			} else {
				p.emit(Event{Kind: "idle"})
			}
		}
	}
	close(p.events)
}

func (p *Player) emit(e Event) {
	select {
	case p.events <- e:
	default: // drop if the consumer is slow; UI only needs the latest
	}
}

func (p *Player) Close() error {
	if p.conn != nil {
		_ = p.conn.Close()
	}
	reap(p.cmd)
	cleanupIPC(p.sock) // platform-specific; removes the socket file on Unix, no-op on Windows
	return nil
}

// reap kills mpv and waits for it, so no zombie process is left behind.
func reap(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	_ = cmd.Process.Kill()
	_, _ = cmd.Process.Wait()
}
