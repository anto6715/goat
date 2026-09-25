package safermapp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/anto6715/goat/internal/manifest"
)

const testMD5 = "0123456789abcdef0123456789abcdef"

func TestIndexRecord(t *testing.T) {
	manifestDir := t.TempDir()

	tests := []struct {
		name     string
		entry    manifest.Entry
		wantErr  bool
		wantName string
		wantMD5  string
		wantPath string
	}{
		{
			name:     "simple filename",
			entry:    manifest.Entry{Name: "photo.jpg", MD5: testMD5},
			wantName: "photo.jpg",
			wantMD5:  testMD5,
			wantPath: filepath.Join(manifestDir, "photo.jpg"),
		},
		{
			name:     "filename with spaces",
			entry:    manifest.Entry{Name: "my photo.jpg", MD5: testMD5},
			wantName: "my photo.jpg",
			wantMD5:  testMD5,
			wantPath: filepath.Join(manifestDir, "my photo.jpg"),
		},
		{
			name:     "uppercase MD5 is normalized",
			entry:    manifest.Entry{Name: "photo.jpg", MD5: strings.ToUpper(testMD5)},
			wantName: "photo.jpg",
			wantMD5:  testMD5,
			wantPath: filepath.Join(manifestDir, "photo.jpg"),
		},
		{
			name:    "empty filename",
			entry:   manifest.Entry{Name: "", MD5: testMD5},
			wantErr: true,
		},
		{
			name:    "parent traversal",
			entry:   manifest.Entry{Name: "../photo.jpg", MD5: testMD5},
			wantErr: true,
		},
		{
			name:    "nested filename",
			entry:   manifest.Entry{Name: filepath.Join("sub", "photo.jpg"), MD5: testMD5},
			wantErr: true,
		},
		{
			name: "cleaned nested filename",
			entry: manifest.Entry{
				Name: strings.Join([]string{"sub", "..", "photo.jpg"}, string(os.PathSeparator)),
				MD5:  testMD5,
			},
			wantErr: true,
		},
		{
			name:    "absolute filename",
			entry:   manifest.Entry{Name: filepath.Join(string(os.PathSeparator), "tmp", "photo.jpg"), MD5: testMD5},
			wantErr: true,
		},
		{
			name:    "current directory",
			entry:   manifest.Entry{Name: ".", MD5: testMD5},
			wantErr: true,
		},
		{
			name:    "empty MD5",
			entry:   manifest.Entry{Name: "photo.jpg", MD5: ""},
			wantErr: true,
		},
		{
			name:    "short MD5",
			entry:   manifest.Entry{Name: "photo.jpg", MD5: "01234567"},
			wantErr: true,
		},
		{
			name:    "non hexadecimal MD5",
			entry:   manifest.Entry{Name: "photo.jpg", MD5: strings.Repeat("z", 32)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, record, err := indexRecord(manifestDir, tt.entry)
			if tt.wantErr {
				if err == nil {
					t.Fatal("indexRecord returned nil error")
				}
				return
			}

			if err != nil {
				t.Fatalf("indexRecord returned an error: %v", err)
			}
			if key.name != tt.wantName || key.md5 != tt.wantMD5 {
				t.Errorf("key = %+v, want name %q and MD5 %q", key, tt.wantName, tt.wantMD5)
			}
			if record.path != tt.wantPath {
				t.Errorf("record path = %q, want %q", record.path, tt.wantPath)
			}
			if record.manifestDir != manifestDir {
				t.Errorf("record manifest directory = %q, want %q", record.manifestDir, manifestDir)
			}
		})
	}
}

func TestAddManifestToIndex(t *testing.T) {
	dir := t.TempDir()
	m := manifest.Manifest{
		Entries: map[string]manifest.Entry{
			"photo.jpg": {
				Name: "photo.jpg",
				MD5:  strings.ToUpper(testMD5),
			},
		},
	}
	if err := manifest.Save(dir, m); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	index := make(fileIndex)
	manifestFile := filepath.Join(dir, manifest.LegacyMetadataFile)
	if err := addManifestToIndex(index, manifestFile); err != nil {
		t.Fatalf("addManifestToIndex returned an error: %v", err)
	}

	key := fileKey{name: "photo.jpg", md5: testMD5}
	records := index[key]
	if len(records) != 1 {
		t.Fatalf("records for key %+v = %d, want 1", key, len(records))
	}
	if records[0].path != filepath.Join(dir, "photo.jpg") {
		t.Errorf("record path = %q, want %q", records[0].path, filepath.Join(dir, "photo.jpg"))
	}
}

func TestAddManifestToIndexRejectsInvalidEntry(t *testing.T) {
	dir := t.TempDir()
	manifestFile := filepath.Join(dir, manifest.LegacyMetadataFile)
	if err := os.WriteFile(
		manifestFile,
		[]byte(strings.Repeat("z", 32)+" photo.jpg\n"),
		0o600,
	); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	err := addManifestToIndex(make(fileIndex), manifestFile)
	if err == nil {
		t.Fatal("addManifestToIndex returned nil error")
	}
	if !strings.Contains(err.Error(), "photo.jpg") {
		t.Errorf("error %q does not contain entry name", err)
	}
	if !strings.Contains(err.Error(), manifestFile) {
		t.Errorf("error %q does not contain manifest path", err)
	}
}
