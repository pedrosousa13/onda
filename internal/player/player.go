package player

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"sync"
	"time"
)

const defaultDialTimeout = 3 * time.Second

// normalizeFilter is mpv's dynaudnorm audio filter tuned for live streams.
// dynaudnorm fills its full Gaussian window (gausssize frames of framelen ms)
// as look-ahead before it emits any audio, so the defaults (f=500, g=31 ≈ 15.5s)
// force ~15s of buffering before playback starts and keep the demuxer cache
// under constant pressure. A 1.1s window (f=100, g=11) cuts startup latency by
// >90% while keeping gain changes gradual enough to avoid pumping.
const normalizeFilter = "dynaudnorm=f=100:g=11"

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

	bin       string // resolved mpv binary name/path
	started   bool
	normalize bool
	volume    *int // desired startup volume, applied via --volume; nil if never set
}

// New verifies mpv exists on PATH but does not start it or touch the audio
// device — mpv is started lazily on the first Play() (see ensureStarted).
func New(opts Options) (*Player, error) {
	bin := opts.Binary
	if bin == "" {
		bin = "mpv"
	}
	if _, err := exec.LookPath(bin); err != nil {
		return nil, errors.New("mpv not found on PATH: install mpv (e.g. `brew install mpv`, or `scoop install mpv` on Windows)")
	}
	p := &Player{
		bin:       bin,
		sock:      ipcAddress(), // platform-specific (Unix socket path or Windows pipe name)
		events:    make(chan Event, 16),
		normalize: opts.Normalize,
	}
	return p, nil
}

// ensureStarted starts mpv and connects its IPC socket, exactly once. Safe to
// call repeatedly/concurrently; only the first call does the work.
func (p *Player) ensureStarted() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.started {
		return nil
	}

	cleanupIPC(p.sock) // remove any stale socket file (no-op on Windows)
	args := []string{"--idle=yes", "--no-video", "--no-terminal", "--input-ipc-server=" + p.sock}
	if p.normalize {
		args = append(args, "--af="+normalizeFilter)
	}
	if p.volume != nil {
		args = append(args, fmt.Sprintf("--volume=%d", *p.volume))
	}
	cmd := exec.Command(p.bin, args...)
	if err := cmd.Start(); err != nil {
		return err
	}
	conn, err := dialWithRetry(p.sock) // platform-specific dialer
	if err != nil {
		reap(cmd)
		return err
	}
	p.cmd = cmd
	p.conn = conn
	if err := p.observeLocked(); err != nil {
		p.closeLocked()
		return err
	}
	p.started = true
	go p.readLoop()
	return nil
}

func (p *Player) Events() <-chan Event { return p.events }

func (p *Player) Play(url string) error {
	if err := p.ensureStarted(); err != nil {
		return err
	}
	return p.send("loadfile", url)
}

func (p *Player) Stop() error {
	if !p.isStarted() {
		return nil
	}
	return p.send("stop")
}

func (p *Player) Pause() error {
	if !p.isStarted() {
		return nil
	}
	return p.send("set_property", "pause", true)
}

func (p *Player) Resume() error {
	if !p.isStarted() {
		return nil
	}
	return p.send("set_property", "pause", false)
}

// Volume sets mpv's volume if already running; otherwise it stores pct as the
// desired startup volume, applied via the --volume launch flag once mpv starts.
func (p *Player) Volume(pct int) error {
	if !p.isStarted() {
		p.mu.Lock()
		p.volume = &pct
		p.mu.Unlock()
		return nil
	}
	return p.send("set_property", "volume", pct)
}

// SetNormalize toggles loudness normalization live by setting mpv's audio-filter
// chain to dynaudnorm (on) or clearing it (off). If mpv hasn't started yet, it
// updates the stored flag so ensureStarted passes/omits --af at launch.
func (p *Player) SetNormalize(on bool) error {
	if !p.isStarted() {
		p.mu.Lock()
		p.normalize = on
		p.mu.Unlock()
		return nil
	}
	if on {
		return p.send("set_property", "af", normalizeFilter)
	}
	return p.send("set_property", "af", "")
}

func (p *Player) isStarted() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.started
}

// observeLocked registers the property observers used to emit title/idle/playing
// events. Callers must already hold p.mu (used from ensureStarted during startup).
func (p *Player) observeLocked() error {
	if err := p.sendLocked("observe_property", 1, "media-title"); err != nil {
		return err
	}
	return p.sendLocked("observe_property", 2, "core-idle")
}

func (p *Player) send(args ...any) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.sendLocked(args...)
}

// sendLocked writes a command to mpv's IPC connection. Callers must hold p.mu.
func (p *Player) sendLocked(args ...any) error {
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
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.started {
		return nil
	}
	p.closeLocked()
	return nil
}

// closeLocked tears down the mpv process and IPC connection. Callers must hold
// p.mu and must only call this when mpv has actually been started.
func (p *Player) closeLocked() {
	if p.conn != nil {
		_ = p.conn.Close()
	}
	reap(p.cmd)
	cleanupIPC(p.sock) // platform-specific; removes the socket file on Unix, no-op on Windows
}

// reap kills mpv and waits for it, so no zombie process is left behind.
func reap(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	_ = cmd.Process.Kill()
	_, _ = cmd.Process.Wait()
}
