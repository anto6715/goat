package find

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.Filter != "*" {
		t.Fatalf("Filter = %q, want %q", opts.Filter, "*")
	}
	if opts.MaxDepth != -1 {
		t.Fatalf("MaxDepth = %d, want %d", opts.MaxDepth, -1)
	}
}

func TestFindFilesDefaultIsUnlimited(t *testing.T) {
	root := writeTestTree(t)

	opts := DefaultOptions()
	opts.Filter = "*.txt"

	paths, err := FindFilesWithOptions(root, opts)
	if err != nil {
		t.Fatalf("FindFiles returned error: %v", err)
	}

	assertRelativePaths(t, root, paths, []string{
		"root.txt",
		"sub/child.txt",
		"sub/nested/grandchild.txt",
	})
}

func TestFindFilesWithOptionsMaxDepthZero(t *testing.T) {
	root := writeTestTree(t)

	paths, err := FindFilesWithOptions(root, Options{
		Filter:   "*.txt",
		MaxDepth: 0,
	})
	if err != nil {
		t.Fatalf("FindFilesWithOptions returned error: %v", err)
	}

	assertRelativePaths(t, root, paths, []string{
		"root.txt",
	})
}

func TestFindFilesWithOptionsMaxDepthOne(t *testing.T) {
	root := writeTestTree(t)

	paths, err := FindFilesWithOptions(root, Options{
		Filter:   "*.txt",
		MaxDepth: 1,
	})
	if err != nil {
		t.Fatalf("FindFilesWithOptions returned error: %v", err)
	}

	assertRelativePaths(t, root, paths, []string{
		"root.txt",
		"sub/child.txt",
	})
}

func TestFindFilesWithOptionsMaxDepthUnlimited(t *testing.T) {
	root := writeTestTree(t)

	paths, err := FindFilesWithOptions(root, Options{
		Filter:   "*.txt",
		MaxDepth: -1,
	})
	if err != nil {
		t.Fatalf("FindFilesWithOptions returned error: %v", err)
	}

	assertRelativePaths(t, root, paths, []string{
		"root.txt",
		"sub/child.txt",
		"sub/nested/grandchild.txt",
	})
}

func writeTestTree(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	files := map[string]string{
		"root.txt":                  "root",
		"root.log":                  "ignore",
		"sub/child.txt":             "child",
		"sub/nested/grandchild.txt": "grandchild",
	}

	for rel, content := range files {
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll(%s): %v", path, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("WriteFile(%s): %v", path, err)
		}
	}

	return root
}

func assertRelativePaths(t *testing.T, root string, got []string, want []string) {
	t.Helper()

	relPaths := make([]string, 0, len(got))
	for _, path := range got {
		rel, err := filepath.Rel(root, path)
		if err != nil {
			t.Fatalf("filepath.Rel(%q, %q): %v", root, path, err)
		}
		relPaths = append(relPaths, filepath.ToSlash(rel))
	}

	slices.Sort(relPaths)
	slices.Sort(want)

	if !slices.Equal(relPaths, want) {
		t.Fatalf("paths = %v, want %v", relPaths, want)
	}
}
