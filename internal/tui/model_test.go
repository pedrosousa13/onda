package tui

import (
	"errors"
	"testing"

	"github.com/pedrosousa13/radio/internal/domain"
)

var errSample = errors.New("boom")

func TestUpdateStationsMsgPopulatesList(t *testing.T) {
	m := Model{}
	updated, _ := m.Update(stationsMsg{stations: []domain.Station{{Name: "KEXP"}}})
	got := updated.(Model)
	if len(got.stations) != 1 || got.stations[0].Name != "KEXP" {
		t.Fatalf("stationsMsg did not populate list: %+v", got.stations)
	}
}

func TestUpdateErrMsgSetsStatus(t *testing.T) {
	m := Model{}
	updated, _ := m.Update(errMsg{err: errSample})
	if got := updated.(Model); got.status == "" {
		t.Fatal("errMsg should set a non-empty status line")
	}
}

type fakeStore struct {
	adds, removes int
	isFav         bool
	custom        []domain.Station
}

func (f *fakeStore) Favorites() ([]domain.Station, error)    { return nil, nil }
func (f *fakeStore) AddFavorite(domain.Station) error        { f.adds++; f.isFav = true; return nil }
func (f *fakeStore) RemoveFavorite(domain.Station) error     { f.removes++; f.isFav = false; return nil }
func (f *fakeStore) IsFavorite(domain.Station) (bool, error) { return f.isFav, nil }
func (f *fakeStore) AddCustom(s domain.Station) error        { f.custom = append(f.custom, s); return nil }
func (f *fakeStore) SaveQuality(domain.QualityPref) error    { return nil }
func (f *fakeStore) SaveTracking(string) error               { return nil }
func (f *fakeStore) SaveHistory(bool) error                  { return nil }

func TestToggleFavoriteAddsAndRemoves(t *testing.T) {
	fs := &fakeStore{}
	m := Model{store: fs, stations: []domain.Station{{Name: "KEXP", Homepage: "kexp.org"}}}
	updated, _ := m.toggleFavorite() // add
	m = updated
	updated, _ = m.toggleFavorite() // remove
	m = updated
	if fs.adds != 1 || fs.removes != 1 {
		t.Fatalf("toggle should add then remove, got adds=%d removes=%d", fs.adds, fs.removes)
	}
}

func TestSettingsCycleQuality(t *testing.T) {
	m := Model{quality: domain.QualityHighest}
	m = m.cycleQuality()
	if m.quality == domain.QualityHighest {
		t.Fatal("cycleQuality should change the value")
	}
}
