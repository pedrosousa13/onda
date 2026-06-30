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

func variants(bitrates ...int) []StreamVariant {
	v := make([]StreamVariant, len(bitrates))
	for i, b := range bitrates {
		v[i] = StreamVariant{URL: "u", Bitrate: b}
	}
	return v
}

func TestSelectVariant(t *testing.T) {
	cases := []struct {
		name string
		pref QualityPref
		in   []int
		want int // expected Bitrate
	}{
		{"highest", QualityHighest, []int{64, 320, 128}, 320},
		{"lowest", QualityLowest, []int{64, 320, 128}, 64},
		{"balanced picks <=128", QualityBalanced, []int{64, 320, 128}, 128},
		{"balanced at exactly 128 (boundary)", QualityBalanced, []int{128, 192}, 128},
		{"balanced falls back to lowest above 128", QualityBalanced, []int{192, 320}, 192},
		{"unknown bitrate treated as 0", QualityHighest, []int{0, 0}, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := Station{Variants: variants(c.in...)}
			got, ok := s.SelectVariant(c.pref)
			if !ok {
				t.Fatal("expected a variant")
			}
			if got.Bitrate != c.want {
				t.Fatalf("want %d, got %d", c.want, got.Bitrate)
			}
		})
	}
}

func TestSelectVariantLossless(t *testing.T) {
	s := Station{Variants: []StreamVariant{
		{URL: "a", Bitrate: 320},
		{URL: "b", Bitrate: 0, Lossless: true}, // HiFi, reported with bitrate 0
		{URL: "c", Bitrate: 128},
	}}
	if v, _ := s.SelectVariant(QualityHighest); !v.Lossless {
		t.Fatalf("highest should pick the lossless variant, got %+v", v)
	}
	if v, _ := s.SelectVariant(QualityLowest); v.Bitrate != 128 || v.Lossless {
		t.Fatalf("lowest should avoid lossless and pick 128, got %+v", v)
	}
}

func TestVariantQualityLabel(t *testing.T) {
	if got := (StreamVariant{Lossless: true}).Quality(); got != "HiFi" {
		t.Fatalf("want HiFi, got %q", got)
	}
	if got := (StreamVariant{Bitrate: 192}).Quality(); got != "192k" {
		t.Fatalf("want 192k, got %q", got)
	}
	if got := (StreamVariant{Codec: "aac"}).Quality(); got != "AAC" {
		t.Fatalf("want AAC, got %q", got)
	}
}

func TestSelectVariantEmpty(t *testing.T) {
	s := Station{}
	if _, ok := s.SelectVariant(QualityHighest); ok {
		t.Fatal("expected ok=false for no variants")
	}
}
