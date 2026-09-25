package safermapp

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateCandidatePresence(t *testing.T) {
	t.Run("regular target and reference", func(t *testing.T) {
		dir := t.TempDir()
		target := writeTestFile(t, filepath.Join(dir, "target", "photo.jpg"))
		reference := writeTestFile(t, filepath.Join(dir, "reference", "photo.jpg"))

		if err := validateCandidatePresence(testCandidate(target, reference)); err != nil {
			t.Fatalf("validateCandidatePresence returned an error: %v", err)
		}
	})

	t.Run("one regular reference is sufficient", func(t *testing.T) {
		dir := t.TempDir()
		target := writeTestFile(t, filepath.Join(dir, "target", "photo.jpg"))
		missingReference := filepath.Join(dir, "missing", "photo.jpg")
		referenceDirectory := makeTestDir(t, filepath.Join(dir, "reference-directory"))
		regularReference := writeTestFile(t, filepath.Join(dir, "reference", "photo.jpg"))

		candidate := testCandidate(
			target,
			missingReference,
			referenceDirectory,
			regularReference,
		)
		if err := validateCandidatePresence(candidate); err != nil {
			t.Fatalf("validateCandidatePresence returned an error: %v", err)
		}
	})

	t.Run("missing target", func(t *testing.T) {
		dir := t.TempDir()
		target := filepath.Join(dir, "missing-target", "photo.jpg")
		reference := writeTestFile(t, filepath.Join(dir, "reference", "photo.jpg"))

		err := validateCandidatePresence(testCandidate(target, reference))
		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("validateCandidatePresence error = %v, want %v", err, os.ErrNotExist)
		}
	})
}

func TestExecutePlanRemovesRegularTargetWithRegularReference(t *testing.T) {
	dir := t.TempDir()
	target := writeTestFile(t, filepath.Join(dir, "target", "photo.jpg"))
	reference := writeTestFile(t, filepath.Join(dir, "reference", "photo.jpg"))

	if err := executePlan([]candidate{testCandidate(target, reference)}); err != nil {
		t.Fatalf("executePlan returned an error: %v", err)
	}

	if _, err := os.Lstat(target); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("target still exists or could not be inspected: %v", err)
	}
	if info, err := os.Lstat(reference); err != nil {
		t.Fatalf("reference was removed or could not be inspected: %v", err)
	} else if !info.Mode().IsRegular() {
		t.Fatalf("reference is no longer a regular file: %v", info.Mode())
	}
}

func TestExecutePlanRefusesUnsafeCandidates(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T) candidate
	}{
		{
			name: "missing reference",
			setup: func(t *testing.T) candidate {
				dir := t.TempDir()
				target := writeTestFile(t, filepath.Join(dir, "target", "photo.jpg"))
				return testCandidate(target, filepath.Join(dir, "missing", "photo.jpg"))
			},
		},
		{
			name: "target directory",
			setup: func(t *testing.T) candidate {
				dir := t.TempDir()
				target := makeTestDir(t, filepath.Join(dir, "target", "photo.jpg"))
				reference := writeTestFile(t, filepath.Join(dir, "reference", "photo.jpg"))
				return testCandidate(target, reference)
			},
		},
		{
			name: "target symbolic link",
			setup: func(t *testing.T) candidate {
				dir := t.TempDir()
				realTarget := writeTestFile(t, filepath.Join(dir, "real-target", "photo.jpg"))
				target := filepath.Join(dir, "target-link")
				if err := os.Symlink(realTarget, target); err != nil {
					t.Fatalf("create target symlink: %v", err)
				}
				reference := writeTestFile(t, filepath.Join(dir, "reference", "photo.jpg"))
				return testCandidate(target, reference)
			},
		},
		{
			name: "reference directory",
			setup: func(t *testing.T) candidate {
				dir := t.TempDir()
				target := writeTestFile(t, filepath.Join(dir, "target", "photo.jpg"))
				reference := makeTestDir(t, filepath.Join(dir, "reference", "photo.jpg"))
				return testCandidate(target, reference)
			},
		},
		{
			name: "reference symbolic link",
			setup: func(t *testing.T) candidate {
				dir := t.TempDir()
				target := writeTestFile(t, filepath.Join(dir, "target", "photo.jpg"))
				realReference := writeTestFile(t, filepath.Join(dir, "real-reference", "photo.jpg"))
				reference := filepath.Join(dir, "reference-link")
				if err := os.Symlink(realReference, reference); err != nil {
					t.Fatalf("create reference symlink: %v", err)
				}
				return testCandidate(target, reference)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plannedCandidate := tt.setup(t)

			if err := executePlan([]candidate{plannedCandidate}); err == nil {
				t.Fatal("executePlan returned nil error")
			}

			if _, err := os.Lstat(plannedCandidate.target.path); err != nil {
				t.Fatalf("unsafe target was removed or could not be inspected: %v", err)
			}
		})
	}
}

func testCandidate(target string, references ...string) candidate {
	referenceRecords := make([]fileRecord, 0, len(references))
	for _, reference := range references {
		referenceRecords = append(referenceRecords, fileRecord{path: reference})
	}

	return candidate{
		key:        fileKey{name: filepath.Base(target), md5: testMD5},
		target:     fileRecord{path: target},
		references: referenceRecords,
	}
}

func writeTestFile(t *testing.T, path string) string {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create parent directory for %q: %v", path, err)
	}
	if err := os.WriteFile(path, []byte("test data"), 0o600); err != nil {
		t.Fatalf("write test file %q: %v", path, err)
	}
	return path
}
