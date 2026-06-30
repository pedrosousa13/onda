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
