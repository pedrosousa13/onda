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
