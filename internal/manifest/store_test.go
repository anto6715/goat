package manifest

import "testing"

func TestSaveReplacesManifest(t *testing.T) {
	dir := t.TempDir()

	oldManifest := Manifest{
		Entries: map[string]Entry{
			"old.txt": {
				Name: "old.txt",
				MD5:  "old-sum",
			},
		},
	}

	if err := Save(dir, oldManifest); err != nil {
		t.Fatalf("save old manifest: %v", err)
	}

	newManifest := Manifest{
		Entries: map[string]Entry{
			"new.txt": {
				Name: "new.txt",
				MD5:  "new-sum",
			},
		},
	}

	if err := Save(dir, newManifest); err != nil {
		t.Fatalf("save new manifest: %v", err)
	}

	got, err := Load(dir)
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}

	if len(got.Entries) != 1 {
		t.Fatalf(
			"expected 1 entry, got %d",
			len(got.Entries),
		)
	}

	if _, exists := got.Entries["old.txt"]; exists {
		t.Fatal("old.txt should have been replaced")
	}

	entry, exists := got.Entries["new.txt"]
	if !exists {
		t.Fatal("new.txt should exist")
	}

	if entry.MD5 != "new-sum" {
		t.Fatalf(
			"expected MD5 %q, got %q",
			"new-sum",
			entry.MD5,
		)
	}
}
