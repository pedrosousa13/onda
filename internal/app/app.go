package app

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pedrosousa13/onda/internal/directory"
	"github.com/pedrosousa13/onda/internal/domain"
	"github.com/pedrosousa13/onda/internal/player"
	"github.com/pedrosousa13/onda/internal/store"
	"github.com/pedrosousa13/onda/internal/tui"
)

var version = "0.1.0-dev"

// Version returns the build version (overridable via -ldflags).
func Version() string { return version }

var rbMirrors = []string{
	"https://de1.api.radio-browser.info",
	"https://nl1.api.radio-browser.info",
	"https://fi1.api.radio-browser.info",
}

func Run() error {
	st, err := store.New()
	if err != nil {
		return err
	}
	cfg, err := st.LoadConfig()
	if err != nil {
		return err
	}

	showFirstRunNoticeOnce(st)

	cacheDir := filepath.Join(cacheRoot(), "onda")
	dir := &directory.Directory{
		Online: directory.NewRadioBrowser(directory.RBOptions{
			Mirrors:   rbMirrors,
			UserAgent: "onda/" + version,
		}),
		Offline: directory.NewOffline(),
		Cache:   directory.NewCache(cacheDir, 24*time.Hour),
	}

	p, err := player.New(player.Options{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "onda needs mpv for playback.")
		fmt.Fprintln(os.Stderr, "Install it (e.g. `brew install mpv`, or `scoop install mpv` on Windows) and try again.")
		return err
	}
	defer p.Close()

	model := tui.New(dir, p, st, domain.QualityPref(cfg.Quality), cfg.Tracking,
		cfg.HistoryEnabled, cfg.Theme, cfg.UpdateCheck, version, cacheDir)
	prog := tea.NewProgram(model, tea.WithAltScreen())

	// Bridge player events into the TUI.
	go func() {
		for e := range p.Events() {
			if e.Kind == "title" {
				prog.Send(tui.TitleMsg(e.Title))
			}
		}
	}()

	_, err = prog.Run()
	return err
}

func cacheRoot() string {
	if d, err := os.UserCacheDir(); err == nil {
		return d
	}
	return os.TempDir()
}

func showFirstRunNoticeOnce(st *store.Store) {
	marker := st.MarkerPath("first-run-shown")
	if _, err := os.Stat(marker); err == nil {
		return
	}
	// Mark it shown up front so the notice appears exactly once — even if the
	// user quits at the prompt rather than pressing Enter.
	_ = os.WriteFile(marker, []byte("1"), 0o644)
	fmt.Println("A quick note on privacy:")
	fmt.Println()
	fmt.Println("onda connects you directly to broadcasters — like opening a stream in a")
	fmt.Println("browser or VLC, they (and, for non-HTTPS streams, your network) can see")
	fmt.Println("what you're playing. Searches go to the public Radio Browser service.")
	fmt.Println("onda itself never records, rebroadcasts, or reports your listening —")
	fmt.Println("popularity tracking is off by default (change it in settings).")
	fmt.Println("onda also checks GitHub once a day for new versions; turn this off in settings.")
	fmt.Println()
	fmt.Println("Shown once. Press Enter to continue — you won't see this again.")
	fmt.Scanln()
}
