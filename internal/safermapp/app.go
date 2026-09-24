package safermapp

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/anto6715/goat/find"
	"github.com/anto6715/goat/internal/manifest"
)

type Options struct {
	Depth int
}

func Run(root string, opts Options) error {
	hashFiles, err := LoadHashFiles(root, opts.Depth)
	if err != nil {
		return err
	}
	fmt.Println("hash files:", len(hashFiles))
	// for _, hashFile := range hashFiles {
		// fmt.Println(hashFile)
	// }
	return nil
}

func LoadHashFiles(root string, depth int) ([]HashFile, error) {
	// get list of all available manifest files
	manifestFiles, err := loadManifest(root, depth)
	if err != nil {
		return nil, err
	}

	// for each manifest file, parse it and extract hash files
	finalHashFiles := make([]HashFile, 0, len(manifestFiles))
	for _, manifestFile := range manifestFiles {
		hashFiles, err := processManifest(manifestFile)
		if err != nil {
			return nil, err
		}
		finalHashFiles = append(finalHashFiles, hashFiles...)
	}
	return finalHashFiles, nil
}

func loadManifest(root string, depth int) ([]string, error) {
	findOpts := find.DefaultOptions()
	findOpts.Filter = manifest.LegacyMetadataFile
	findOpts.MaxDepth = depth
	slog.Info("finding manifest", "root", root, "depth", depth)
	return find.FindFilesWithOptions(root, findOpts)
}

func processManifest(manifestFile string) ([]HashFile, error) {
	slog.Info("parsing manifest", "file", manifestFile)
	manifestDir, err := filepath.Abs(filepath.Dir(manifestFile))
	if err != nil {
		return nil, err
	}

	manifest, err := manifest.Load(manifestDir)
	if err != nil {
		return nil, err
	}

	slog.Debug("manifest parsed", "file", manifestFile, "entries", len(manifest.Entries))

	hashFiles := make([]HashFile, 0, len(manifest.Entries))
	for _, entry := range manifest.Entries {
		hashFiles = append(hashFiles, HashFile{
			Hash: entry.MD5,
			Path: filepath.Join(manifestDir, entry.Name),
		})
	}
	return hashFiles, nil
}
