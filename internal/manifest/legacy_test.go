package manifest

import (
	"strings"
	"testing"
)

func TestReadLegacy(t *testing.T) {
	input := strings.NewReader(
		"03d23540721e4aebb4624297be70c270 go.mod\n" +
			"1086735daabd30618c808c9328781b47 go.sum",
	)

	m, err := ReadLegacy(input)
	if err != nil {
		t.Fatal(err)
	}

	entry := m.Entries["go.mod"]
	if entry.MD5 != "03d23540721e4aebb4624297be70c270" {
		t.Fatalf("expected MD5 %q, got %q", "03d23540721e4aebb4624297be70c270", entry.MD5)
	}

	if len(m.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(m.Entries))
	}
}

func TestWriteLegacy(t *testing.T) {
	// defined unsorted entries
	m := Manifest{
		Entries: map[string]Entry{
			"go.sum": {
				Name: "go.sum",
				MD5:  "1086735daabd30618c808c9328781b47",
			},
			"go.mod": {
				Name: "go.mod",
				MD5:  "03d23540721e4aebb4624297be70c270",
			},
		},
	}

	w := strings.Builder{}
	if err := WriteLegacy(&w, m); err != nil {
		t.Fatal(err)
	}

	// verify the output is sorted
	expected := "03d23540721e4aebb4624297be70c270 go.mod\n1086735daabd30618c808c9328781b47 go.sum\n"
	if w.String() != expected {
		t.Fatalf("expected %q, got %q", expected, w.String())
	}
}
