package store

// Config holds user preferences (persisted as TOML).
type Config struct {
	Quality        string `toml:"quality"`         // highest|lowest|balanced
	Tracking       string `toml:"tracking"`        // never|opt-in|opt-out
	HistoryEnabled bool   `toml:"history_enabled"`
}

func DefaultConfig() Config {
	return Config{Quality: "highest", Tracking: "never", HistoryEnabled: false}
}
