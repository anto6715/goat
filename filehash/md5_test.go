package filehash

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMD5Sum(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.txt")

	if err := os.WriteFile(path, []byte("hello world"), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}

	got, err := MD5Sum(path)
	if err != nil {
		t.Fatalf("MD5Sum(%s) returned error: %v", path, err)
	}

	const want = "5eb63bbbe01eeed093cb22bb8f5acdc3"
	if got != want {
		t.Fatalf("MD5Sum(%s) = %q, want %q", path, got, want)
	}
}
