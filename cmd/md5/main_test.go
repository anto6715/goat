package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/anto6715/goat/filehash"
)

func TestRunHashesFilesInWalkOrder(t *testing.T) {
	dir := t.TempDir()

	files := []struct {
		name    string
		content string
	}{
		{name: "a.txt", content: "alpha"},
		{name: "b.txt", content: "bravo"},
		{name: "c.txt", content: "charlie"},
	}

	var wantLines []string
	for _, file := range files {
		path := filepath.Join(dir, file.name)
		if err := os.WriteFile(path, []byte(file.content), 0o600); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}

		sum, err := filehash.MD5Sum(path)
		if err != nil {
			t.Fatalf("hash %s: %v", path, err)
		}

		wantLines = append(wantLines, fmt.Sprintf("%s %s", sum, path))
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := run(cli{Path: dir, NWorker: 2}, &stdout, &stderr)
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	if got := strings.TrimSpace(stderr.String()); got != "" {
		t.Fatalf("stderr = %q, want empty", got)
	}

	want := strings.Join(wantLines, "\n")
	if got := strings.TrimSpace(stdout.String()); got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestRunRejectsInvalidWorkerCount(t *testing.T) {
	dir := t.TempDir()

	err := run(cli{Path: dir, NWorker: 0}, io.Discard, io.Discard)
	if err == nil {
		t.Fatal("expected error for invalid worker count")
	}

	if got := err.Error(); got != "n-worker must be greater than 0" {
		t.Fatalf("error = %q, want %q", got, "n-worker must be greater than 0")
	}
}

func TestRunReturnsErrorWhenAFileCannotBeHashed(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission-based hashing failure test is Unix-specific")
	}

	dir := t.TempDir()
	goodPath := filepath.Join(dir, "good.txt")
	badPath := filepath.Join(dir, "bad.txt")

	if err := os.WriteFile(goodPath, []byte("good"), 0o600); err != nil {
		t.Fatalf("write %s: %v", goodPath, err)
	}
	if err := os.WriteFile(badPath, []byte("bad"), 0o600); err != nil {
		t.Fatalf("write %s: %v", badPath, err)
	}
	if err := os.Chmod(badPath, 0o000); err != nil {
		t.Fatalf("chmod %s: %v", badPath, err)
	}
	defer func() {
		_ = os.Chmod(badPath, 0o600)
	}()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err := run(cli{Path: dir, NWorker: 2}, &stdout, &stderr)
	if !errors.Is(err, errHashFailed) {
		t.Fatalf("error = %v, want %v", err, errHashFailed)
	}

	if strings.Contains(stdout.String(), badPath) {
		t.Fatalf("stdout should not contain failed path %q: %q", badPath, stdout.String())
	}
	if !strings.Contains(stdout.String(), goodPath) {
		t.Fatalf("stdout should contain successful path %q: %q", goodPath, stdout.String())
	}
	if !strings.Contains(stderr.String(), badPath) {
		t.Fatalf("stderr should contain failed path %q: %q", badPath, stderr.String())
	}
}
