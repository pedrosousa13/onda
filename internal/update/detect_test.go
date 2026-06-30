package update

import (
	"path/filepath"
	"testing"
)

func TestClassifyPath(t *testing.T) {
	cases := []struct{ path, want string }{
		{"/opt/homebrew/Cellar/onda/1.0.0/bin/onda", "homebrew"},
		{"/usr/local/Cellar/onda/1.0.0/bin/onda", "homebrew"},
		{"/home/linuxbrew/.linuxbrew/bin/onda", "homebrew"},
		{`C:\Users\me\scoop\apps\onda\current\onda.exe`, "scoop"},
		{"/usr/local/bin/onda", "other"},
		{"/home/me/.local/bin/onda", "other"},
	}
	for _, c := range cases {
		if got := classifyPath(c.path); got != c.want {
			t.Errorf("classifyPath(%q)=%q want %q", c.path, got, c.want)
		}
	}
}

func TestDirWritable(t *testing.T) {
	dir := t.TempDir()
	if !dirWritable(dir) {
		t.Fatal("temp dir should be writable")
	}
	if dirWritable(filepath.Join(dir, "does-not-exist")) {
		t.Fatal("nonexistent dir should not be writable")
	}
}
