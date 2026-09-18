package manifest

import (
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
