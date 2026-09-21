package manifest

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Public Methods ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
func Save(root string, m Manifest) error {
	return withExclusiveLock(root, func() error {
		return save(root, m)
	})
}

// No need (for now) to use exclusive lock for loading
func Load(root string) (Manifest, error) {
	return load(root)
}

// Atomic transaction from loading, merging, and saving
func Update(dir string, other Manifest) error {
	return withExclusiveLock(dir, func() error {
		finalManifest, err := load(dir)
		if err != nil {
			return err
		}
		finalManifest.Merge(other)

		return save(dir, finalManifest)
	})
}

// Private Methods ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
func save(root string, m Manifest) error {
	targetPath := filepath.Join(root, LegacyMetadataFile)

	tmpFile, err := os.CreateTemp(root, LegacyMetadataFile+".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	tmpPath := tmpFile.Name()

	// If anything fails, remove the temporary file
	// If rename works, Remove does nothing
	defer func() {
		_ = os.Remove(tmpPath)
	}()

	// Write temporary file
	if err := WriteLegacy(tmpFile, m); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("failed to write metadata: %w", err)
	}

	// Sync and Close temporary file
	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("failed to sync temp file: %w", err)
	}
	// Close and Save temporary file
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Finally, rename the temp file to the target path
	if err := os.Rename(tmpPath, targetPath); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

func load(root string) (Manifest, error) {
	path := filepath.Join(root, LegacyMetadataFile)
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return Manifest{}, nil
	}
	if err != nil {
		return Manifest{
			Entries: make(map[string]Entry),
		}, fmt.Errorf("failed to open metadata file: %w", err)
	}
	defer file.Close()

	m, err := ReadLegacy(file)
	if err != nil {
		return Manifest{}, fmt.Errorf("failed to read metadata: %w", err)
	}

	return m, nil
}
