package manifest

import "testing"

func TestMerge(t *testing.T) {
	current := Manifest{
		Entries: map[string]Entry{
			"a.txt": {Name: "a.txt", MD5: "old-a"},
			"b.txt": {Name: "b.txt", MD5: "old-b"},
		},
	}
	incoming := Manifest{
		Entries: map[string]Entry{
			"a.txt": {Name: "a.txt", MD5: "new-a"},
			"c.txt": {Name: "c.txt", MD5: "new-c"},
		},
	}

	current.Merge(incoming)

	if len(current.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(current.Entries))
	}

	tests := []struct {
		name    string
		wantMD5 string
	}{
		{name: "a.txt", wantMD5: "new-a"},
		{name: "b.txt", wantMD5: "old-b"},
		{name: "c.txt", wantMD5: "new-c"},
	}

	for _, test := range tests {
		entry, exists := current.Entries[test.name]
		if !exists {
			t.Errorf("expected entry %q to exist", test.name)
			continue
		}
		if entry.MD5 != test.wantMD5 {
			t.Errorf("entry %q: expected MD5 %q, got %q", test.name, test.wantMD5, entry.MD5)
		}
	}
}
