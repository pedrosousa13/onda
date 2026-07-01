package domain

import (
	"testing"
)

func TestAxisStringCountry(t *testing.T) {
	if got := AxisCountry.String(); got != "country" {
		t.Errorf("AxisCountry.String() = %q; want %q", got, "country")
	}
}

func TestAxisStringTag(t *testing.T) {
	if got := AxisTag.String(); got != "tag" {
		t.Errorf("AxisTag.String() = %q; want %q", got, "tag")
	}
}

func TestAxisStringLanguage(t *testing.T) {
	if got := AxisLanguage.String(); got != "language" {
		t.Errorf("AxisLanguage.String() = %q; want %q", got, "language")
	}
}

func TestSortDescendingVotes(t *testing.T) {
	s := Sort{Key: SortVotes, Flip: false}
	if !s.Descending() {
		t.Errorf("Sort{Key: SortVotes}.Descending() = false; want true")
	}
}

func TestSortDescendingVotesFlipped(t *testing.T) {
	s := Sort{Key: SortVotes, Flip: true}
	if s.Descending() {
		t.Errorf("Sort{Key: SortVotes, Flip: true}.Descending() = true; want false")
	}
}

func TestSortDescendingName(t *testing.T) {
	s := Sort{Key: SortName, Flip: false}
	if s.Descending() {
		t.Errorf("Sort{Key: SortName}.Descending() = true; want false")
	}
}

func TestSortDescendingNameFlipped(t *testing.T) {
	s := Sort{Key: SortName, Flip: true}
	if !s.Descending() {
		t.Errorf("Sort{Key: SortName, Flip: true}.Descending() = false; want true")
	}
}

func TestSortDescendingTrend(t *testing.T) {
	s := Sort{Key: SortTrend, Flip: false}
	if !s.Descending() {
		t.Errorf("Sort{Key: SortTrend}.Descending() = false; want true")
	}
}

func TestSortDescendingTrendFlipped(t *testing.T) {
	s := Sort{Key: SortTrend, Flip: true}
	if s.Descending() {
		t.Errorf("Sort{Key: SortTrend, Flip: true}.Descending() = true; want false")
	}
}

func TestSortLabelVotes(t *testing.T) {
	s := Sort{Key: SortVotes, Flip: false}
	if got := s.Label(); got != "votes ↓" {
		t.Errorf("Sort{Key: SortVotes}.Label() = %q; want %q", got, "votes ↓")
	}
}

func TestSortLabelVotesFlipped(t *testing.T) {
	s := Sort{Key: SortVotes, Flip: true}
	if got := s.Label(); got != "votes ↑" {
		t.Errorf("Sort{Key: SortVotes, Flip: true}.Label() = %q; want %q", got, "votes ↑")
	}
}

func TestSortLabelName(t *testing.T) {
	s := Sort{Key: SortName, Flip: false}
	if got := s.Label(); got != "name ↑" {
		t.Errorf("Sort{Key: SortName}.Label() = %q; want %q", got, "name ↑")
	}
}

func TestSortLabelTrending(t *testing.T) {
	s := Sort{Key: SortTrend, Flip: false}
	if got := s.Label(); got != "trending ↓" {
		t.Errorf("Sort{Key: SortTrend}.Label() = %q; want %q", got, "trending ↓")
	}
}

func TestSortNext(t *testing.T) {
	tests := []struct {
		name  string
		input Sort
		want  Sort
	}{
		{
			name:  "SortVotes cycles to SortName",
			input: Sort{Key: SortVotes, Flip: true},
			want:  Sort{Key: SortName, Flip: false},
		},
		{
			name:  "SortName cycles to SortTrend",
			input: Sort{Key: SortName, Flip: true},
			want:  Sort{Key: SortTrend, Flip: false},
		},
		{
			name:  "SortTrend cycles to SortVotes",
			input: Sort{Key: SortTrend, Flip: true},
			want:  Sort{Key: SortVotes, Flip: false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.Next()
			if got != tt.want {
				t.Errorf("Sort.Next() = %+v; want %+v", got, tt.want)
			}
		})
	}
}

func TestStationLanguageAndTrend(t *testing.T) {
	s := Station{
		Name:       "Test Radio",
		Country:    "Portugal",
		Language:   "pt",
		Votes:      42,
		ClickCount: 100,
		Trend:      5,
	}
	if s.Language != "pt" {
		t.Errorf("Station.Language = %q; want %q", s.Language, "pt")
	}
	if s.Trend != 5 {
		t.Errorf("Station.Trend = %d; want %d", s.Trend, 5)
	}
}
