package directory

import (
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/pedrosousa13/onda/internal/domain"
)

// corpusSchema is bumped whenever the cached shape or grouping logic changes,
// so an old dump becomes a miss and is refetched.
const corpusSchema = "v2"

// CorpusStore persists the full station dump on disk (gzip), with a TTL.
type CorpusStore struct {
	path string
	ttl  time.Duration
	now  func() time.Time
}

func NewCorpusStore(dir string, ttl time.Duration) *CorpusStore {
	return &CorpusStore{
		path: filepath.Join(dir, "stations-"+corpusSchema+".json.gz"),
		ttl:  ttl,
		now:  time.Now,
	}
}

type corpusFile struct {
	FetchedAt time.Time        `json:"fetched_at"`
	Stations  []domain.Station `json:"stations"`
}

// Load reads and decodes the cached dump. ok is false on missing/corrupt/empty.
func (s *CorpusStore) Load() ([]domain.Station, time.Time, bool) {
	f, err := os.Open(s.path)
	if err != nil {
		return nil, time.Time{}, false
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, time.Time{}, false
	}
	defer gz.Close()
	var cf corpusFile
	if err := json.NewDecoder(gz).Decode(&cf); err != nil || len(cf.Stations) == 0 {
		return nil, time.Time{}, false
	}
	return cf.Stations, cf.FetchedAt, true
}

// Save writes the dump atomically (temp file + rename) so an interrupted write
// never leaves a broken corpus.
func (s *CorpusStore) Save(stations []domain.Station) error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "corpus-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op once renamed away

	gz := gzip.NewWriter(tmp)
	if err := json.NewEncoder(gz).Encode(corpusFile{FetchedAt: s.now(), Stations: stations}); err != nil {
		gz.Close()
		tmp.Close()
		return err
	}
	if err := gz.Close(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, s.path)
}

// Fresh reports whether a dump fetched at fetchedAt is still within the TTL.
func (s *CorpusStore) Fresh(fetchedAt time.Time) bool {
	return s.now().Sub(fetchedAt) <= s.ttl
}

// Delete removes the cached dump. A missing file is not an error.
func (s *CorpusStore) Delete() error {
	if err := os.Remove(s.path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Size returns the on-disk dump size in bytes, and whether it exists.
func (s *CorpusStore) Size() (int64, bool) {
	fi, err := os.Stat(s.path)
	if err != nil {
		return 0, false
	}
	return fi.Size(), true
}
