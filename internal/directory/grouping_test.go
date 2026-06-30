package directory

import "testing"

func TestGroupRecordsMergesBitrates(t *testing.T) {
	recs := []record{
		{Name: "KEXP", Homepage: "kexp.org", URL: "u64", Bitrate: 64},
		{Name: "KEXP", Homepage: "kexp.org", URL: "u128", Bitrate: 128},
		{Name: "Other", Homepage: "x.com", URL: "uX", Bitrate: 96},
	}
	got := GroupRecords(recs)
	if len(got) != 2 {
		t.Fatalf("want 2 stations, got %d", len(got))
	}
	var kexp int = -1
	for i, s := range got {
		if s.Name == "KEXP" {
			kexp = i
		}
	}
	if kexp == -1 || len(got[kexp].Variants) != 2 {
		t.Fatalf("KEXP should have 2 variants, got %+v", got)
	}
}

func TestGroupRecordsCaseInsensitiveKey(t *testing.T) {
	recs := []record{
		{Name: "kexp", Homepage: "KEXP.org", URL: "a", Bitrate: 64},
		{Name: "KEXP", Homepage: "kexp.org", URL: "b", Bitrate: 128},
	}
	if len(GroupRecords(recs)) != 1 {
		t.Fatal("case-different names+homepages should merge")
	}
}

func TestGroupRecordsCollapsesFIPFamily(t *testing.T) {
	// All of these are the same station ("FIP", France) in Radio Browser.
	recs := []record{
		{Name: "FIP", Country: "France", URL: "1", Bitrate: 192},
		{Name: "FIP", Country: "France", URL: "2", Bitrate: 192},
		{Name: "FIP (hifi.aac)", Country: "France", URL: "3"},
		{Name: "FIP (metadata)", Country: "France", URL: "4", Bitrate: 192},
		{Name: "FIP (no pub)", Country: "France", URL: "5"},
		{Name: "FIP aac", Country: "France", URL: "6", Bitrate: 192},
		{Name: "Fip", Country: "France", URL: "7", Bitrate: 192},
	}
	got := GroupRecords(recs)
	if len(got) != 1 {
		names := make([]string, len(got))
		for i, s := range got {
			names[i] = s.Name
		}
		t.Fatalf("FIP family should collapse to 1 station, got %d: %v", len(got), names)
	}
	if got[0].Name != "FIP" {
		t.Fatalf("display name should be 'FIP', got %q", got[0].Name)
	}
	// After dedupe by quality label the 7 records collapse to distinct qualities:
	// 192k (several), HiFi (hifi.aac), and "—" (no-codec bitrate-0).
	if len(got[0].Variants) != 3 {
		labels := make([]string, len(got[0].Variants))
		for i, v := range got[0].Variants {
			labels[i] = v.Quality()
		}
		t.Fatalf("expected 3 distinct-quality variants, got %d: %v", len(got[0].Variants), labels)
	}
}

func TestGroupRecordsMergesQualitySuffixes(t *testing.T) {
	// Real FIP-style duplicates: same station, split by quality/format suffixes
	// and punctuation, with differing homepages.
	recs := []record{
		{Name: "FIP Jazz", Country: "France", Homepage: "a", URL: "u1", Bitrate: 192},
		{Name: "FIP Jazz (Hi-Fi)", Country: "France", Homepage: "b", URL: "u2", Bitrate: 0},
		{Name: "FIP Jazz (hifi.aac)", Country: "France", Homepage: "c", URL: "u3", Bitrate: 0},
		{Name: "FIP Hip Hop", Country: "France", URL: "u4", Bitrate: 192},
		{Name: "FIP Hip-Hop", Country: "France", URL: "u5", Bitrate: 107},
	}
	got := GroupRecords(recs)
	if len(got) != 2 {
		t.Fatalf("want 2 stations (FIP Jazz, FIP Hip Hop), got %d: %+v", len(got), got)
	}
	var jazz *struct{ n int }
	for i := range got {
		if got[i].Name == "FIP Jazz" {
			jazz = &struct{ n int }{i}
		}
	}
	if jazz == nil {
		t.Fatalf("expected a station named 'FIP Jazz', got %+v", got)
	}
	j := got[jazz.n]
	// 192k + two HiFi records → after dedupe: HiFi, 192k (2 distinct qualities).
	if len(j.Variants) != 2 {
		t.Fatalf("FIP Jazz should have 2 distinct-quality variants, got %d", len(j.Variants))
	}
	// Best-first: the HiFi variant should sort to the front.
	if !j.Variants[0].Lossless {
		t.Fatalf("expected lossless variant first, got %+v", j.Variants[0])
	}
}
