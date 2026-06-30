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

func TestSelectVariantEmpty(t *testing.T) {
	s := Station{}
	if _, ok := s.SelectVariant(QualityHighest); ok {
		t.Fatal("expected ok=false for no variants")
	}
}
