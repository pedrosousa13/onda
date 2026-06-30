package app

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pedrosousa13/radio/internal/directory"
	"github.com/pedrosousa13/radio/internal/domain"
	"github.com/pedrosousa13/radio/internal/player"
	"github.com/pedrosousa13/radio/internal/store"
	"github.com/pedrosousa13/radio/internal/tui"
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

	cacheDir := filepath.Join(cacheRoot(), "radio")
	dir := &directory.Directory{
		Online: directory.NewRadioBrowser(directory.RBOptions{
			Mirrors:   rbMirrors,
			UserAgent: "radio/" + version,
		}),
		Offline: directory.NewOffline(),
		Cache:   directory.NewCache(cacheDir, 24*time.Hour),
	}

	p, err := player.New(player.Options{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "radio needs mpv for playback.")
		fmt.Fprintln(os.Stderr, "Install it (e.g. `brew install mpv`, or `scoop install mpv` on Windows) and try again.")
		return err
	}
	defer p.Close()

	model := tui.New(dir, p, st, domain.QualityPref(cfg.Quality), cfg.Tracking, cfg.HistoryEnabled)
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
	fmt.Println("radio streams directly from broadcasters — they see your IP,")
	fmt.Println("and searches go to the public Radio Browser service. radio never")
	fmt.Println("records, rebroadcasts, or reports what you listen to (by default).")
	fmt.Println("Press Enter to continue.")
	fmt.Scanln()
	_ = os.WriteFile(marker, []byte("1"), 0o644)
}
