package safermapp

import (
	"encoding/hex"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/anto6715/goat/find"
	"github.com/anto6715/goat/internal/manifest"
)

func loadManifest(root string, depth int) ([]string, error) {
	findOpts := find.DefaultOptions()
	findOpts.Filter = manifest.LegacyMetadataFile
	findOpts.MaxDepth = depth
	slog.Info("finding manifest", "root", root, "depth", depth)
	return find.FindFilesWithOptions(root, findOpts)
}

func processManifests(index fileIndex, manifestFiles []string) error {
	for _, manifestFile := range manifestFiles {
		err := addManifestToIndex(index, manifestFile)
		if err != nil {
			return err
		}
	}
	return nil
}

func addManifestToIndex(index fileIndex, manifestFile string) error {
	// Parsing and Load
	slog.Info("parsing manifest", "file", manifestFile)
	manifestDir, err := getManifestDir(manifestFile)
	if err != nil {
		return fmt.Errorf("failed to get manifest dir: %w", err)
	}

	m, err := manifest.Load(manifestDir)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}
	slog.Debug("manifest parsed", "file", manifestFile, "entries", len(m.Entries))

	// Manifest Indexing
	for _, entry := range m.Entries {
		key, record, err := indexRecord(manifestDir, entry)
		if err != nil {
			return fmt.Errorf("invalid entry %q in manifest %q: %w", entry.Name, manifestFile, err)
		}
		index[key] = append(index[key], record)
	}
	return nil
}

func indexRecord(manifestDir string, entry manifest.Entry) (fileKey, fileRecord, error) {

	err := isValidEntry(entry)
	if err != nil {
		return fileKey{}, fileRecord{}, err
	}

	normalizedMD5 := strings.ToLower(entry.MD5)
	key := fileKey{
		name: entry.Name,
		md5:  normalizedMD5,
	}
	record := fileRecord{
		path:        filepath.Join(manifestDir, entry.Name),
		manifestDir: manifestDir,
	}
	return key, record, nil
}

func isValidEntry(entry manifest.Entry) error {
	// Check on file name
	if entry.Name == "" {
		return fmt.Errorf("empty entry name")
	}
	if !filepath.IsLocal(entry.Name) {
		return fmt.Errorf("not local entry: %s", entry.Name)
	}
	if entry.Name == "." || entry.Name == ".." {
		return fmt.Errorf("invalid entry: %s", entry.Name)
	}
	if filepath.Base(entry.Name) != entry.Name {
		return fmt.Errorf("entry must be a filename: %s", entry.Name)
	}

	// Check on MD5
	if entry.MD5 == "" {
		return fmt.Errorf("empty entry MD5")
	}

	if len(entry.MD5) != 32 {
		return fmt.Errorf("invalid MD5 length: %d", len(entry.MD5))
	}

	if _, err := hex.DecodeString(entry.MD5); err != nil {
		return fmt.Errorf("invalid MD5: %w", err)
	}

	return nil
}
