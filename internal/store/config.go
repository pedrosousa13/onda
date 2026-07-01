package store

// Config holds user preferences (persisted as TOML).
type Config struct {
	Quality        string `toml:"quality"`  // highest|lowest|balanced
	Tracking       string `toml:"tracking"` // never|opt-in|opt-out
	HistoryEnabled bool   `toml:"history_enabled"`
	Theme          string `toml:"theme"`           // e.g. catppuccin-mocha
	UpdateCheck    bool   `toml:"update_check"`    // daily check for new releases (opt-out)
	LiveSearch     bool   `toml:"live_search"`     // search as you type; off → enter-to-search
	Volume         int    `toml:"volume"`          // last playback volume, 0–100
	Normalize      bool   `toml:"normalize"`       // even out loudness across stations (opt-in)
	OfflineCatalog string `toml:"offline_catalog"` // ask|on|off — download full station dump
}

func DefaultConfig() Config {
	return Config{Quality: "highest", Tracking: "never", HistoryEnabled: false, Theme: "catppuccin-mocha", UpdateCheck: true, LiveSearch: true, Volume: 100, Normalize: false, OfflineCatalog: "ask"}
}
