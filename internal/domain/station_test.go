package domain

import "testing"

func TestStationHasVariants(t *testing.T) {
	s := Station{
		Name:    "KEXP",
		Country: "United States",
		Variants: []StreamVariant{
			{URL: "http://a/64", Codec: "MP3", Bitrate: 64},
			{URL: "http://a/128", Codec: "MP3", Bitrate: 128},
		},
	}
	if len(s.Variants) != 2 {
		t.Fatalf("want 2 variants, got %d", len(s.Variants))
	}
}
