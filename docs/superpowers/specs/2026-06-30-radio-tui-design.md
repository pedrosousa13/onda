# Spec: `radio` — an ethical, privacy-minded radios-of-the-world TUI

**Date:** 2026-06-30
**Status:** Approved (design), pending implementation plan
**Repo:** https://github.com/pedrosousa13/radio (public)

## Summary

A single-binary terminal application that lets you wander the world by radio:
drill down through places, search, add your own stations, and listen — streaming
**directly from the broadcaster**, with nothing hosted or rebroadcast by us. It
captures the spirit of Radio Garden in the terminal while being legal, ethical,
and privacy-minded by default.

## Goals

- Discover internet radio stations geographically (Continent → Country → Station).
- Fast keyboard-driven search (name / country / tag / language).
- Favorites and user-added custom stations as **required** first-class features.
- Auto-pick stream quality (highest bitrate by default; configurable; per-play override).
- Work on first run and offline via a bundled public-domain station list.
- Be silent-by-default about listening behavior; lead with ethics and privacy.
- Run on **Linux, macOS, and Windows** as first-class targets.
- Install in one command (Homebrew on macOS/Linux, Scoop on Windows); ship as a portable single binary per OS.

## Non-goals (explicitly out of scope)

- **Recording / ripping streams** — never (reproduction is a copyright concern).
- **Geo-restriction circumvention** — never (no built-in proxy/VPN bypass).
- **Rebroadcast / proxy / re-hosting** of any stream — never.
- **Self-hosting any server or database** — the app is a pure client.
- ASCII world map — deferred to a later release (see "Future").
- Xiph/Icecast directory source — deferred (requires HTML scraping).
- Scraping Radio Garden's private API — rejected (no reuse grant; ethically off).

## Legal & ethical model

The app behaves as a **user agent** (like a browser or VLC): it opens a stream URL
the user chose, connecting their machine **directly** to the broadcaster's own
public server. Each listener is a single direct connection the broadcaster invited
by publishing the stream. Responsibility for any given stream's legality rests with
the broadcaster, not the client.

Three invariants keep the app inside that model:

1. **No recording** of streams to disk.
2. **No geo-circumvention** — geo-blocked streams fail gracefully.
3. **No rebroadcast/proxy** — one user, one direct connection.

Audio decoding is delegated to a system backend (`mpv`), so codec licensing
(AAC, etc.) is handled at the system level and stays out of this codebase. MP3
patents expired in 2017.

## Data sources

- **Primary — Radio Browser** (`radio-browser.info`): community directory, data is
  **public domain**, **no API key**, **nothing to host**. Accessed via the public
  DNS mirror pool `all.api.radio-browser.info` with fallback across mirrors. The
  client sends an **honest descriptive `User-Agent`** (e.g. `radio/<version>`) and
  **caches results locally**.
- **Bundled offline fallback** — the `deroverda/recommended-radio-streams` list
  (**CC0**, ~300 curated stations), embedded in the binary so the app works on
  first run and with no network. CC0 = safe to embed and redistribute.

Radio Browser responses are **cached locally** as JSON under the XDG data dir
(see `store`) with a modest TTL (~24h) so repeat use is fast and offline-tolerant;
stale cache is still served when the network or all mirrors are unreachable.

**Important data constraint:** Radio Browser stores **one stream URL per record**.
A station offering multiple bitrates appears as **multiple separate records**.
The `directory` layer is responsible for **grouping records into one logical
station** with multiple stream variants, which is what makes quality selection
possible.

**Rejected sources** (with reasons): Shoutcast (partner-gated key, no-storage +
anti-copyleft license, privacy callback); Radio Garden (no reuse grant, ethically
off); SomaFM API (officially closed to third parties); RadioDNS (a hybrid-radio
standard, not a directory); Dirble (dead since 2019); Xiph dir (stream URLs require
HTML scraping — possible later).

## Privacy posture

- **Popularity tracking is silent by default.** Radio Browser exposes `/click` and
  `/vote` endpoints that report which station you played. The app exposes a setting
  with three modes — `never` / `opt-in` / `opt-out` — and **defaults to `never`**.
  Out of the box the client reports nothing about what you listen to.
- **Honest first-run notice**, shown once: streaming connects you directly to
  broadcasters, who see your IP (inherent to all radio apps); and directory
  searches go to Radio Browser mirrors. Stated plainly, no dark patterns.
- **All user data is local-only** (favorites, custom stations, config, optional
  history). Play history is **optional** and easy to clear or disable.
- **No telemetry or analytics** of any kind in the app itself.

## Stack

- **Language/UI:** Go 1.24+ + Bubble Tea (Elm architecture) + Lip Gloss (styling).
  Rationale: beautiful TUIs with modest effort, **single static binary**, strong
  HTTP/concurrency for streaming, excellent cross-compilation and distribution.
- **Audio:** a **headless `mpv` subprocess** controlled over its **JSON IPC channel**
  (a Unix domain socket on Linux/macOS, a named pipe on Windows via `go-winio`).
  Rationale: every codec/format/HLS works, robust, ICY metadata support, and it
  keeps codec licensing out of our code. `mpv` is a declared install dependency.
- **Cross-platform:** Linux, macOS, Windows. Only the `player` IPC connection is
  platform-split (build-tagged files); `store` paths use Go's `os.UserConfigDir()`/
  `os.UserCacheDir()`, which already resolve per-OS.

## Architecture

Five focused packages, each with a single responsibility and a clean interface:

| Package | Responsibility | Key interface (illustrative) |
|---|---|---|
| `directory` | Fetch station data; group multi-bitrate records into one logical `Station` | `Search(query)`, `Countries()`, `StationsBy(filter)` → `[]Station` |
| `player` | Audio playback via a headless `mpv` subprocess (JSON IPC) | `Play(url)`, `Stop()`, `Pause()`, `Volume(n)`, `Events() <-chan Event` |
| `store` | Local persistence in XDG dirs: config, favorites, custom stations, optional history | file read/write — config as TOML, data (favorites/custom/history/cache) as JSON |
| `tui` | Bubble Tea models + Lip Gloss styling: browser, search, now-playing, favorites, add-station form, settings | Elm-style `Update`/`View` |
| `app` | Wiring + lifecycle: spawn/teardown `mpv`, graceful shutdown, dependency wiring | `main` |

`directory` has **two backends behind one interface** — a Radio Browser HTTP client
(mirror-pool selection with fallback, honest User-Agent, local cache) and the
embedded CC0 list — plus a merge/group layer producing the domain model below.

## Domain model

- **`Station`** (logical): `{ name, country, geo, tags, homepage, variants []StreamVariant }`
- **`StreamVariant`**: `{ url, codec, bitrate, hls }`

Grouping Radio Browser's one-URL-per-record entries into a `Station` with multiple
`variants` is the `directory` layer's job. Auto-quality selects the highest-bitrate
variant by default; the setting offers **lowest** (the minimum-bitrate variant) and
**balanced** (the highest variant at or below ~128 kbps, else the lowest available).
The user can override the variant per play. Custom user-added stations use the same
model and grouping.

## Data flow

1. The TUI dispatches async `tea.Cmd`s for directory queries so the UI **never blocks**.
2. `directory` returns `[]Station`; list models update.
3. Selecting a station resolves the preferred `StreamVariant` and calls `player.Play(url)`.
4. `player` emits state and ICY-metadata events; the now-playing bar updates.
5. Favorites, custom stations, and config are read/written through `store`.

## Features (v1, required unless noted)

- Drill-down browser: Continent → Country → Station, with counts and live filtering.
- Search across name / country / tag / language.
- Play / stop / pause / volume; now-playing bar with live ICY track metadata.
- **Favorites** (required): star any station from any source; local; dedicated view;
  favorited Radio Browser stations cache their stream URL for resilience.
- **Custom links** (required): add a stream URL with a name (+ optional country/tags)
  via an in-app form; saved locally; listed alongside results; favoritable; works
  fully offline; participates in quality grouping.
- Auto-quality: default highest bitrate; setting for lowest / balanced; per-play override.
- Online (Radio Browser) + embedded CC0 offline list.
- Settings: quality default, popularity-tracking mode (never/opt-in/opt-out), history toggle.
- Local-only persistence with the privacy defaults above.

## Error handling

- Network/API failure → roll across Radio Browser mirrors → fall back to cache, then
  the embedded list → show a non-blocking toast. The app stays usable offline.
- Dead or geo-blocked stream, or `mpv` playback error → "stream unavailable"; the user
  picks another. Never crash; never attempt a bypass.
- `mpv` not installed → detected at startup with a clear, actionable install hint.

## Distribution

- **GoReleaser** cross-compiles (macOS arm64/amd64, Linux), publishes **GitHub
  Releases**, and **auto-generates the Homebrew formula**.
- **macOS/Linux — Homebrew via a custom tap** — `pedrosousa13/homebrew-tap`:
  `brew install pedrosousa13/tap/radio`. The formula declares **`depends_on "mpv"`**,
  so one command yields a working app with its audio backend. (homebrew-core needs
  notability; a tap is the right call now.)
- **Windows — Scoop via a custom bucket** — `pedrosousa13/scoop-bucket`:
  `scoop install radio`. Scoop can't force the `mpv` dependency the way Homebrew
  does, so the README instructs `scoop install mpv`; the app also prints a clear
  "install mpv" message if it's missing.
- **Alternative installs** (any OS) documented: `go install`, and direct
  prebuilt-binary download from Releases.
- The **exact** `brew` command, tap URL, and version are finalized when the release
  pipeline is set up, so the README's install commands match what actually ships
  rather than being invented ahead of time.

## Testing strategy

- `directory`: unit tests over recorded Radio Browser JSON fixtures (grouping +
  quality-pick logic); a fake HTTP server to exercise mirror fallback.
- `player`: unit-test the IPC control protocol against a mock socket; one integration
  test that a real `mpv` plays a local audio fixture.
- `store`: round-trip read/write tests in a temp XDG directory.
- `tui`: Bubble Tea `Update` is a pure function — send messages, assert model state.

## Future (post-v1)

- ASCII world-map view as an optional discovery mode (toggle alongside the browser).
- Optional Xiph/Icecast directory source.
- Asciinema demo / screenshots in the README.

## Open items finalized at implementation time

- Final project/binary name (currently `radio`, matching the repo and working dir).
- Exact Homebrew tap URL and first release version (drive the README install commands).
- Concrete keybinding map (drafted in the README, confirmed during UI build).

---

## Appendix: README draft

> This reflects the agreed design. Install commands marked `«finalized at release»`
> are filled in accurately when the release pipeline exists; they are not invented here.

```markdown
# radio

Wander the world by radio, from your terminal. `radio` is an ethical,
privacy-minded internet-radio TUI in the spirit of Radio Garden: drill down
through places, search, add your own stations, and listen — streaming directly
from broadcasters.

## Ethics & privacy

`radio` behaves like a browser pointed at a stream you chose. It connects you
**directly** to the broadcaster's public server.

- **No recording**, ever.
- **No geo-unblocking** — region-locked streams simply fail.
- **No proxy, no rebroadcast** — one direct connection, nothing re-hosted.
- **Silent by default** — it does not report what you listen to. A setting lets
  you opt in to contributing to community popularity rankings if you want.
- **Local-only data** — favorites, custom stations, and config never leave your
  machine. No telemetry.
- **Public-domain data** — station data comes from the public-domain Radio Browser
  project and a bundled CC0 station list.

Streaming inherently exposes your IP to the broadcaster, and searches go to Radio
Browser mirrors — the same as any internet-radio app. `radio` tells you this on
first run.

## Install

`radio` runs on Linux, macOS, and Windows. It requires [`mpv`](https://mpv.io)
on your PATH for playback.

### macOS / Linux — Homebrew (recommended)

```sh
brew install «finalized at release: pedrosousa13/tap/radio»
```

This also installs `mpv`.

### Windows — Scoop

```sh
scoop bucket add pedrosousa13 https://github.com/pedrosousa13/scoop-bucket
scoop install radio
scoop install mpv
```

### Alternatives (any OS)

- `go install «finalized at release: module path»@latest` (install `mpv` separately)
- Download the matching archive from the [Releases](https://github.com/pedrosousa13/radio/releases) page (ensure `mpv` is on PATH)

## Usage

Launch:

```sh
radio
```

- **Browse:** move through Continent → Country → Station. Type to filter the
  current list.
- **Search:** search by name, country, tag, or language.
- **Play:** select a station to start playing. The now-playing bar shows the
  current track when the station provides it.
- **Favorites:** star a station to save it; open the Favorites view anytime.
- **Add your own:** open the add-station form, paste a stream URL, give it a name.
- **Quality:** `radio` auto-picks the highest bitrate; override it per play or
  change the default in settings.

Keybindings are shown in-app (drafted during the UI build).

## Configuration

Config and data live under your XDG directories (e.g. `~/.config/radio` and
`~/.local/share/radio`):

- **Quality default** — highest (default) / lowest / balanced
- **Popularity tracking** — never (default) / opt-in / opt-out
- **History** — optional; clear or disable at any time

## Data sources & credits

- [Radio Browser](https://www.radio-browser.info) — public-domain station directory
- [`deroverda/recommended-radio-streams`](https://github.com/deroverda/recommended-radio-streams) — CC0 bundled station list
- [`mpv`](https://mpv.io) — audio playback backend

## License

TBD (chosen before first release).
```
