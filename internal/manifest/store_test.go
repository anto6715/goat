package manifest

import (
	"fmt"
	"sync"
	"testing"
)

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

func TestUpdateMergesManifest(t *testing.T) {
	dir := t.TempDir()

	existing := Manifest{
		Entries: map[string]Entry{
			"existing.txt": {Name: "existing.txt", MD5: "existing-sum"},
			"changed.txt":  {Name: "changed.txt", MD5: "old-sum"},
		},
	}
	if err := Save(dir, existing); err != nil {
		t.Fatalf("save existing manifest: %v", err)
	}

	incoming := Manifest{
		Entries: map[string]Entry{
			"changed.txt": {Name: "changed.txt", MD5: "new-sum"},
			"new.txt":     {Name: "new.txt", MD5: "new-file-sum"},
		},
	}
	if err := Update(dir, incoming); err != nil {
		t.Fatalf("update manifest: %v", err)
	}

	got, err := Load(dir)
	if err != nil {
		t.Fatalf("load updated manifest: %v", err)
	}

	want := map[string]string{
		"existing.txt": "existing-sum",
		"changed.txt":  "new-sum",
		"new.txt":      "new-file-sum",
	}
	if len(got.Entries) != len(want) {
		t.Fatalf("expected %d entries, got %d", len(want), len(got.Entries))
	}

	for name, wantMD5 := range want {
		entry, exists := got.Entries[name]
		if !exists {
			t.Errorf("expected entry %q to exist", name)
			continue
		}
		if entry.MD5 != wantMD5 {
			t.Errorf("entry %q: expected MD5 %q, got %q", name, wantMD5, entry.MD5)
		}
	}
}

func TestConcurrentUpdatesPreserveAllEntries(t *testing.T) {
	dir := t.TempDir()
	const updateCount = 20

	start := make(chan struct{})
	errors := make(chan error, updateCount)

	var wg sync.WaitGroup
	wg.Add(updateCount)

	for i := range updateCount {
		go func() {
			defer wg.Done()
			<-start

			name := fmt.Sprintf("file-%02d.txt", i)
			sum := fmt.Sprintf("sum-%02d", i)
			incoming := Manifest{
				Entries: map[string]Entry{
					name: {Name: name, MD5: sum},
				},
			}

			if err := Update(dir, incoming); err != nil {
				errors <- fmt.Errorf("update %q: %w", name, err)
			}
		}()
	}

	close(start)
	wg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}
	if t.Failed() {
		return
	}

	got, err := Load(dir)
	if err != nil {
		t.Fatalf("load manifest after concurrent updates: %v", err)
	}

	if len(got.Entries) != updateCount {
		t.Fatalf("expected %d entries, got %d", updateCount, len(got.Entries))
	}

	for i := range updateCount {
		name := fmt.Sprintf("file-%02d.txt", i)
		wantMD5 := fmt.Sprintf("sum-%02d", i)

		entry, exists := got.Entries[name]
		if !exists {
			t.Errorf("expected entry %q to exist", name)
			continue
		}
		if entry.MD5 != wantMD5 {
			t.Errorf("entry %q: expected MD5 %q, got %q", name, wantMD5, entry.MD5)
		}
	}
}
