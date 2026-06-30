package tui

import "testing"

func TestSanitizeTitle(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"plain title passes through", "Khruangbin - Maria También", "Khruangbin - Maria También"},
		{"dalet xml → artist — title",
			`<?xml version="1.0"?><RadioInfo><Table><DB_DALET_ARTIST_NAME>Sam</DB_DALET_ARTIST_NAME><DB_DALET_TITLE_NAME>Tentação</DB_DALET_TITLE_NAME></Table></RadioInfo>`,
			"Sam — Tentação"},
		{"unknown markup dropped", `<weird><blob>no useful field</blob></weird>`, ""},
		{"empty stays empty", "", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := sanitizeTitle(c.in); got != c.want {
				t.Fatalf("sanitizeTitle(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}
