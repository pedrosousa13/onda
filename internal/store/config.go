package store

// Config holds user preferences (persisted as TOML).
type Config struct {
	Quality        string `toml:"quality"`         // highest|lowest|balanced
	Tracking       string `toml:"tracking"`        // never|opt-in|opt-out
	HistoryEnabled bool   `toml:"history_enabled"`
	Theme          string `toml:"theme"`           // e.g. catppuccin-mocha
	UpdateCheck    bool   `toml:"update_check"`    // daily check for new releases (opt-out)
}

func DefaultConfig() Config {
	return Config{Quality: "highest", Tracking: "never", HistoryEnabled: false, Theme: "catppuccin-mocha", UpdateCheck: true}
}
