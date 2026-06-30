// Package store provides local-only persistence under XDG directories.
package store

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Store struct{ dir string }

// New returns a Store rooted at the user's XDG config dir for radio.
func New() (*Store, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(base, "radio")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	return &Store{dir: dir}, nil
}

func (s *Store) configPath() string { return filepath.Join(s.dir, "config.toml") }

func (s *Store) LoadConfig() (Config, error) {
	c := DefaultConfig()
	b, err := os.ReadFile(s.configPath())
	if os.IsNotExist(err) {
		return c, nil
	}
	if err != nil {
		return c, err
	}
	if err := toml.Unmarshal(b, &c); err != nil {
		return c, err
	}
	return c, nil
}

func (s *Store) SaveConfig(c Config) error {
	f, err := os.Create(s.configPath())
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(c)
}
