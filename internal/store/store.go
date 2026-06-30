// Package store provides local-only persistence under XDG directories.
package store

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/pedrosousa13/onda/internal/domain"
)

type Store struct{ dir string }

// New returns a Store rooted at the user's XDG config dir for onda.
func New() (*Store, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(base, "onda")
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

func stationKey(s domain.Station) string {
	return s.Name + "|" + s.Homepage
}

func (s *Store) readList(name string) ([]domain.Station, error) {
	b, err := os.ReadFile(filepath.Join(s.dir, name))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var list []domain.Station
	if err := json.Unmarshal(b, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func (s *Store) writeList(name string, list []domain.Station) error {
	b, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.dir, name), b, 0o644)
}

func (s *Store) Favorites() ([]domain.Station, error) { return s.readList("favorites.json") }

func (s *Store) AddFavorite(st domain.Station) error {
	list, err := s.Favorites()
	if err != nil {
		return err
	}
	for _, e := range list {
		if stationKey(e) == stationKey(st) {
			return nil // already present
		}
	}
	return s.writeList("favorites.json", append(list, st))
}

func (s *Store) RemoveFavorite(st domain.Station) error {
	list, err := s.Favorites()
	if err != nil {
		return err
	}
	out := list[:0]
	for _, e := range list {
		if stationKey(e) != stationKey(st) {
			out = append(out, e)
		}
	}
	return s.writeList("favorites.json", out)
}

func (s *Store) CustomStations() ([]domain.Station, error) { return s.readList("custom.json") }

func (s *Store) AddCustom(st domain.Station) error {
	list, err := s.CustomStations()
	if err != nil {
		return err
	}
	return s.writeList("custom.json", append(list, st))
}

// recentsCap bounds the locally-stored play history.
const recentsCap = 50

func (s *Store) Recents() ([]domain.Station, error) { return s.readList("recents.json") }

// AddRecent prepends st to the play history, most-recent-first, de-duplicated by
// station key and capped at recentsCap. Stored as plain JSON in the config dir
// like favorites — portable and local-only. Callers gate this on the user's
// opt-in history setting.
func (s *Store) AddRecent(st domain.Station) error {
	list, err := s.Recents()
	if err != nil {
		return err
	}
	out := make([]domain.Station, 0, len(list)+1)
	out = append(out, st)
	for _, e := range list {
		if stationKey(e) != stationKey(st) {
			out = append(out, e)
		}
	}
	if len(out) > recentsCap {
		out = out[:recentsCap]
	}
	return s.writeList("recents.json", out)
}

// ClearRecents deletes the play-history file.
func (s *Store) ClearRecents() error {
	err := os.Remove(filepath.Join(s.dir, "recents.json"))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (s *Store) MarkerPath(name string) string { return filepath.Join(s.dir, "."+name) }

func (s *Store) IsFavorite(st domain.Station) (bool, error) {
	favs, err := s.Favorites()
	if err != nil {
		return false, err
	}
	for _, e := range favs {
		if stationKey(e) == stationKey(st) {
			return true, nil
		}
	}
	return false, nil
}

func (s *Store) SaveQuality(q domain.QualityPref) error {
	c, err := s.LoadConfig()
	if err != nil {
		return err
	}
	c.Quality = string(q)
	return s.SaveConfig(c)
}

func (s *Store) SaveTracking(t string) error {
	c, err := s.LoadConfig()
	if err != nil {
		return err
	}
	c.Tracking = t
	return s.SaveConfig(c)
}

func (s *Store) SaveHistory(h bool) error {
	c, err := s.LoadConfig()
	if err != nil {
		return err
	}
	c.HistoryEnabled = h
	return s.SaveConfig(c)
}

func (s *Store) SaveUpdateCheck(v bool) error {
	c, err := s.LoadConfig()
	if err != nil {
		return err
	}
	c.UpdateCheck = v
	return s.SaveConfig(c)
}

func (s *Store) SaveLiveSearch(v bool) error {
	c, err := s.LoadConfig()
	if err != nil {
		return err
	}
	c.LiveSearch = v
	return s.SaveConfig(c)
}

func (s *Store) SaveTheme(theme string) error {
	c, err := s.LoadConfig()
	if err != nil {
		return err
	}
	c.Theme = theme
	return s.SaveConfig(c)
}

func (s *Store) SaveVolume(v int) error {
	c, err := s.LoadConfig()
	if err != nil {
		return err
	}
	c.Volume = v
	return s.SaveConfig(c)
}

func (s *Store) SaveNormalize(v bool) error {
	c, err := s.LoadConfig()
	if err != nil {
		return err
	}
	c.Normalize = v
	return s.SaveConfig(c)
}
