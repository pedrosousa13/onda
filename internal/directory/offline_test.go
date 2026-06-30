package directory

import (
	"context"
	"testing"
)

func TestOfflineListLoads(t *testing.T) {
	src := NewOffline()
	stations, err := src.Search(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(stations) == 0 {
		t.Fatal("embedded offline list is empty")
	}
	if stations[0].Name == "" {
		t.Fatal("station missing name")
	}
}
