package update

import "testing"

func TestIsNewer(t *testing.T) {
	cases := []struct {
		current, latest string
		want            bool
	}{
		{"1.0.0", "v1.0.1", true},
		{"1.0.0", "v1.0.0", false},
		{"1.2.0", "v1.1.9", false},
		{"1.0.0", "1.0.1", true},   // latest missing v prefix
		{"v1.0.0", "v1.0.1", true}, // current already has v
		{"1.0.0", "garbage", false},
		{"1.0.0", "", false},
	}
	for _, c := range cases {
		if got := isNewer(c.current, c.latest); got != c.want {
			t.Errorf("isNewer(%q,%q)=%v want %v", c.current, c.latest, got, c.want)
		}
	}
}
