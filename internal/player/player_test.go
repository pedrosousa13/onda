package player

import "testing"

func TestEncodeCommand(t *testing.T) {
	got, err := encodeCommand(7, "loadfile", "http://x/stream")
	if err != nil {
		t.Fatal(err)
	}
	want := `{"command":["loadfile","http://x/stream"],"request_id":7}` + "\n"
	if string(got) != want {
		t.Fatalf("want %q got %q", want, got)
	}
}

func TestParseLineEvent(t *testing.T) {
	f, err := parseLine([]byte(`{"event":"property-change","name":"media-title","data":"Now Playing"}`))
	if err != nil {
		t.Fatal(err)
	}
	if f.Event != "property-change" || f.Name != "media-title" || f.Data != "Now Playing" {
		t.Fatalf("unexpected frame: %+v", f)
	}
}

func TestParseLineIgnoresReplies(t *testing.T) {
	f, err := parseLine([]byte(`{"error":"success","request_id":7}`))
	if err != nil {
		t.Fatal(err)
	}
	if f.Event != "" {
		t.Fatalf("reply should have empty Event, got %q", f.Event)
	}
}

func TestParseLineEndFileReason(t *testing.T) {
	f, err := parseLine([]byte(`{"event":"end-file","reason":"error"}`))
	if err != nil {
		t.Fatal(err)
	}
	if f.Event != "end-file" || f.Reason != "error" {
		t.Fatalf("unexpected frame: %+v", f)
	}
}

func TestNewRequiresMpv(t *testing.T) {
	// With a bogus binary name, New must fail fast and not leak a process.
	_, err := New(Options{Binary: "definitely-not-mpv-xyz"})
	if err == nil {
		t.Fatal("expected error when mpv binary is missing")
	}
}
