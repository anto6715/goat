package safermapp

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestPathWithin(t *testing.T) {
	root := t.TempDir()
	parent := filepath.Join(root, "parent")
	child := filepath.Join(parent, "child")
	sibling := filepath.Join(root, "sibling")

	tests := []struct {
		name   string
		parent string
		child  string
		want   bool
	}{
		{name: "direct child", parent: parent, child: child, want: true},
		{name: "deeper descendant", parent: parent, child: filepath.Join(child, "nested"), want: true},
		{name: "same path", parent: parent, child: parent, want: false},
		{name: "sibling", parent: parent, child: sibling, want: false},
		{name: "ancestor", parent: child, child: parent, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pathWithin(tt.parent, tt.child)
			if err != nil {
				t.Fatalf("pathWithin returned an error: %v", err)
			}
			if got != tt.want {
				t.Errorf("pathWithin(%q, %q) = %v, want %v", tt.parent, tt.child, got, tt.want)
			}
		})
	}
}

func TestNormalizeConfigCanonicalizesWithoutMutatingInput(t *testing.T) {
	root := t.TempDir()
	reference := makeTestDir(t, filepath.Join(root, "reference"))
	target := makeTestDir(t, filepath.Join(root, "target"))
	referenceLink := filepath.Join(root, "reference-link")
	if err := os.Symlink(reference, referenceLink); err != nil {
		t.Fatalf("create reference symlink: %v", err)
	}

	cfg := Config{
		References: []string{referenceLink},
		Target:     target,
		MaxDepth:   2,
	}
	originalReferences := slices.Clone(cfg.References)

	got, err := normalizeConfig(cfg)
	if err != nil {
		t.Fatalf("normalizeConfig returned an error: %v", err)
	}

	if got.Target != target {
		t.Errorf("normalized target = %q, want %q", got.Target, target)
	}
	if len(got.References) != 1 || got.References[0] != reference {
		t.Errorf("normalized references = %q, want [%q]", got.References, reference)
	}
	if !slices.Equal(cfg.References, originalReferences) {
		t.Errorf("normalizeConfig mutated input references: got %q, want %q", cfg.References, originalReferences)
	}
}

func TestNormalizeConfigRejectsEqualOrOverlappingRoots(t *testing.T) {
	root := t.TempDir()
	parent := makeTestDir(t, filepath.Join(root, "parent"))
	child := makeTestDir(t, filepath.Join(parent, "child"))
	parentLink := filepath.Join(root, "parent-link")
	if err := os.Symlink(parent, parentLink); err != nil {
		t.Fatalf("create parent symlink: %v", err)
	}

	tests := []struct {
		name      string
		reference string
		target    string
	}{
		{name: "equal", reference: parent, target: parent},
		{name: "equal through symlink", reference: parentLink, target: parent},
		{name: "target inside reference", reference: parent, target: child},
		{name: "reference inside target", reference: child, target: parent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := normalizeConfig(Config{
				References: []string{tt.reference},
				Target:     tt.target,
				MaxDepth:   -1,
			})
			if err == nil {
				t.Fatal("normalizeConfig returned nil error")
			}
		})
	}
}

func TestNormalizeConfigAcceptsSiblingRoots(t *testing.T) {
	root := t.TempDir()
	reference := makeTestDir(t, filepath.Join(root, "reference"))
	target := makeTestDir(t, filepath.Join(root, "target"))

	_, err := normalizeConfig(Config{
		References: []string{reference},
		Target:     target,
		MaxDepth:   -1,
	})
	if err != nil {
		t.Fatalf("normalizeConfig returned an error for sibling roots: %v", err)
	}
}

func TestNormalizeConfigRejectsInvalidValues(t *testing.T) {
	validDir := t.TempDir()

	tests := []struct {
		name string
		cfg  Config
	}{
		{
			name: "max depth below unlimited sentinel",
			cfg:  Config{References: []string{validDir}, Target: validDir, MaxDepth: -2},
		},
		{
			name: "no references",
			cfg:  Config{Target: validDir, MaxDepth: -1},
		},
		{
			name: "empty target",
			cfg:  Config{References: []string{validDir}, MaxDepth: -1},
		},
		{
			name: "empty reference",
			cfg:  Config{References: []string{""}, Target: validDir, MaxDepth: -1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := normalizeConfig(tt.cfg); err == nil {
				t.Fatal("normalizeConfig returned nil error")
			}
		})
	}
}

func makeTestDir(t *testing.T, path string) string {
	t.Helper()

	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("create directory %q: %v", path, err)
	}
	return path
}
