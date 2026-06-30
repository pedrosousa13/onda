package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pedrosousa13/onda/internal/domain"
	"github.com/pedrosousa13/onda/internal/update"
)

type stationsMsg struct{ stations []domain.Station }
type errMsg struct{ err error }
type titleMsg struct{ title string }
type updateMsg struct{ status update.Status }
type updateAppliedMsg struct{ err error }

// updateCheckCmd runs the (cached) update check off the UI goroutine.
func updateCheckCmd(current, cacheDir string) tea.Cmd {
	return func() tea.Msg {
		st, _ := update.Check(context.Background(), current, cacheDir)
		return updateMsg{status: st}
	}
}

// applyUpdateCmd performs the self-update off the UI goroutine.
func applyUpdateCmd(st update.Status) tea.Cmd {
	return func() tea.Msg {
		return updateAppliedMsg{err: update.Apply(context.Background(), st)}
	}
}

// playback state events bridged from the player.
type playingMsg struct{}
type idleMsg struct{}
type playErrMsg struct{ err error }

// connectTimeoutMsg fires after a play attempt; attempt guards against stale ticks.
type connectTimeoutMsg struct{ attempt int }

// searchDebounceMsg fires after typing pauses; seq guards against stale ticks.
type searchDebounceMsg struct{ seq int }

// sleepTickMsg fires once a minute while a sleep timer is set; seq guards
// against stale ticks left over from a cancelled or restarted timer.
type sleepTickMsg struct{ seq int }

// searchCmd runs a directory search off the UI goroutine.
func searchCmd(d Searcher, query string) tea.Cmd {
	return func() tea.Msg {
		stations, err := d.Search(context.Background(), query)
		if err != nil {
			return errMsg{err: err}
		}
		return stationsMsg{stations: stations}
	}
}

// popularCmd loads the top-voted stations off the UI goroutine.
func popularCmd(d Searcher) tea.Cmd {
	return func() tea.Msg {
		stations, err := d.Popular(context.Background())
		if err != nil {
			return errMsg{err: err}
		}
		return stationsMsg{stations: stations}
	}
}

// Searcher is the slice of directory the TUI needs (keeps tui decoupled).
type Searcher interface {
	Search(ctx context.Context, query string) ([]domain.Station, error)
	Popular(ctx context.Context) ([]domain.Station, error)
}

// TitleMsg builds a titleMsg from outside the package (used by the app event bridge).
func TitleMsg(s string) tea.Msg { return titleMsg{title: s} }

// Playback event constructors for the app event bridge.
func PlayingMsg() tea.Msg          { return playingMsg{} }
func IdleMsg() tea.Msg             { return idleMsg{} }
func PlayErrMsg(err error) tea.Msg { return playErrMsg{err: err} }
