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
	if len(j.Variants) != 3 {
		t.Fatalf("FIP Jazz should have 3 variants, got %d", len(j.Variants))
	}
	// Best-first: the HiFi variant should sort to the front.
	if !j.Variants[0].Lossless {
		t.Fatalf("expected lossless variant first, got %+v", j.Variants[0])
	}
}
