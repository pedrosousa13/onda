# radio

Wander the world by radio, from your terminal. `radio` is an ethical,
privacy-minded internet-radio TUI in the spirit of Radio Garden: browse stations
by place, search, add your own, and listen — streaming **directly from
broadcasters**, with nothing hosted, recorded, or rebroadcast by us.

> Status: early v1. The bundled offline station list is a small starter set
> pending a full public-domain import (see [Data sources](#data-sources--credits)).

## Ethics & privacy

`radio` behaves like a browser pointed at a stream you chose. It connects you
**directly** to the broadcaster's public server.

- **No recording**, ever.
- **No geo-unblocking** — region-locked streams simply fail.
- **No proxy, no rebroadcast** — one direct connection, nothing re-hosted.
- **Silent by default** — popularity tracking defaults to `never`, so `radio`
  reports nothing about what you listen to. You can opt in (`opt-in`/`opt-out`)
  in settings to contribute to community rankings if you want.
- **Local-only data** — favorites, custom stations, and config never leave your
  machine. No telemetry.
- **Public-domain data** — station data comes from the public-domain
  [Radio Browser](https://www.radio-browser.info) project plus a bundled CC0 list.

Streaming inherently exposes your IP to the broadcaster, and searches go to Radio
Browser mirrors — the same as any internet-radio app. `radio` tells you this on
first run.

## Install

`radio` runs on **Linux, macOS, and Windows**. It requires
[`mpv`](https://mpv.io) on your `PATH` for playback.

### macOS / Linux — Homebrew (recommended)

```sh
brew install pedrosousa13/tap/radio
```

This also installs `mpv`.

### Windows — Scoop

```sh
scoop bucket add pedrosousa13 https://github.com/pedrosousa13/scoop-bucket
scoop install radio
scoop install mpv
```

### Any OS — Go

```sh
go install github.com/pedrosousa13/radio@latest
```

Install `mpv` separately (`brew install mpv`, `scoop install mpv`, or your
distro's package manager).

### Any OS — prebuilt binary

Download the archive for your platform from the
[Releases](https://github.com/pedrosousa13/radio/releases) page and put `radio`
on your `PATH`. Ensure `mpv` is installed.

## Usage

Launch:

```sh
radio
```

### Keys

**Browsing / favorites**

| Key | Action |
|-----|--------|
| `↑`/`↓` or `k`/`j` | Move selection |
| `enter` | Play selected station |
| `s` | Stop |
| `+` / `-` | Volume up / down |
| `f` | Toggle favorite on selected station |
| `F` | Show favorites |
| `p` | Popular (top-voted worldwide) |
| `esc` | Back to Home |
| `/` | Search |
| `a` | Add a custom station |
| `,` | Settings |
| `esc` | Back to browse |
| `q` | Quit |

**Search** — type a query, `enter` to search, `esc` to cancel. Matches **name,
country, and tags** (queried in parallel) and ranks results best-match-first with
light fuzzy/typo tolerance.

**Add a station** — `tab` (or `↑`/`↓`) to move between *name*, *URL*, and
*bitrate* (optional); `enter` to save; `esc` to cancel. Custom stations are saved
locally and appear alongside everything else.

**Settings** — `1` cycles audio quality (highest / balanced / lowest), `2` cycles
popularity tracking (never / opt-in / opt-out), `3` toggles play history, `4` cycles
the **theme**; `esc` to go back. Changes are saved immediately.

When a station offers multiple bitrates, `radio` auto-picks per your quality
setting (default: highest).

On launch you land on **Home** — your now-playing panel plus your favorites (or a
**Popular** preview, the top-voted stations worldwide, until you've saved any).
From anywhere: `esc` returns Home, `p` opens the full Popular list, `F` favorites,
`/` search. Popular comes from Radio Browser's open ranking — reading it reports
nothing about you.

### Themes

Switch in settings (`,` then `4`). Bundled: **Catppuccin** (Mocha, Macchiato,
Frappé, Latte), **Dracula**, **Nord**, **Gruvbox**. Default is Catppuccin Mocha;
your choice is saved to `config.toml`.

## Configuration

Config and data live under your OS config directory (resolved via Go's
`os.UserConfigDir`):

- Linux: `~/.config/radio/`
- macOS: `~/Library/Application Support/radio/`
- Windows: `%AppData%\radio\`

Files:

- `config.toml` — `quality` (highest|balanced|lowest), `tracking`
  (never|opt-in|opt-out), `history_enabled`, `theme`
- `favorites.json`, `custom.json` — your favorites and added stations

Cached Radio Browser results live under your OS cache directory
(`os.UserCacheDir`, e.g. `~/.cache/radio/` on Linux) with a 24-hour TTL; stale
cache is still served when you're offline.

Defaults are privacy-first: quality `highest`, tracking `never`, history disabled.

## On duplicate stations

Radio Browser is crowd-sourced and has **no canonical station IDs**, so the same
station often appears many times (different bitrates, `(Hi-Fi)`/`(metadata)`
suffixes, broadcaster prefixes like `RTP `, punctuation). `radio` merges these
heuristically — normalizing names and grouping by name + country, then offering
the distinct bitrates as a `[ ]` chooser. It's best-effort: a few stragglers may
remain rather than risk merging genuinely different stations. Perfect grouping
isn't possible without canonical IDs the source doesn't provide.

## Data sources & credits

- [Radio Browser](https://www.radio-browser.info) — public-domain station directory
- [`deroverda/recommended-radio-streams`](https://github.com/deroverda/recommended-radio-streams) — CC0 bundled station list
- [`mpv`](https://mpv.io) — audio playback backend
- [Charm](https://charm.sh) — Bubble Tea & Lip Gloss, the TUI foundation

## Building from source

Requires Go 1.24+ and `mpv`.

```sh
git clone https://github.com/pedrosousa13/radio
cd radio
go build -o radio .
./radio
```

Run the tests with `go test ./...` (add `-race` if your Go has a C toolchain).

## License

[MIT](LICENSE) © 2026 Pedro Sousa
