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
