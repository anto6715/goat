package manifest

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func Save(root string, m Manifest) error {
	path := filepath.Join(root, LegacyMetadataFile)

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}

	writeErr := WriteLegacy(file, m)
	closeErr := file.Close()
	if writeErr != nil {
		return fmt.Errorf("failed to write metadata: %w", writeErr)
	}
	if closeErr != nil {
		return fmt.Errorf("failed to close metadata file: %w", closeErr)
	}

	return nil
}

func Load(root string) (Manifest, error) {
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
