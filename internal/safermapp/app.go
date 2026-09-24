package safermapp

import (
	"log/slog"
	"path/filepath"

	"github.com/anto6715/goat/find"
	"github.com/anto6715/goat/internal/manifest"
)

type Config struct {
	References []string
	Target     string
	MaxDepth   int
}

func Run(cfg Config) error {
	err := validateConfig(cfg)
	if err != nil {
		return err
	}

	slog.Info("Loading target files to be removed", "target", cfg.Target)
	targetHashFiles, err := LoadHashFiles(cfg.Target, cfg.MaxDepth)
	if err != nil {
		return err
	}
	slog.Info("target hash files", "nfiles", targetHashFiles.NFiles())

	for _, reference := range cfg.References {
		slog.Info("loading reference hash files", "reference", reference)
		refHashFiles, err := LoadHashFiles(reference, cfg.MaxDepth)
		if err != nil {
			return err
		}
		slog.Info("reference hash files", "nfiles", refHashFiles.NFiles())

		cleanDuplicates(refHashFiles, targetHashFiles)
	}
	return nil

}

// Given a root directory and depth, load all manifest files and
// build a HashFiles containing the hash files from all manifest files found.
func LoadHashFiles(root string, depth int) (HashFiles, error) {
	// get list of all available manifest files
	manifestFiles, err := loadManifest(root, depth)
	if err != nil {
		return HashFiles{}, err
	}

	// for each manifest file, parse it and extract hash files
	finalHashFiles := NewHashFiles()
	for _, manifestFile := range manifestFiles {
		hashFiles, err := processManifest(manifestFile)
		if err != nil {
			return HashFiles{}, err
		}
		finalHashFiles.Merge(hashFiles)
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

func processManifest(manifestFile string) (HashFiles, error) {
	slog.Info("parsing manifest", "file", manifestFile)
	manifestDir, err := filepath.Abs(filepath.Dir(manifestFile))
	if err != nil {
		return HashFiles{}, err
	}

	manifest, err := manifest.Load(manifestDir)
	if err != nil {
		return HashFiles{}, err
	}

	slog.Debug("manifest parsed", "file", manifestFile, "entries", len(manifest.Entries))

	hashFiles := NewHashFiles()
	for _, entry := range manifest.Entries {
		hashFiles.AddPath(entry.MD5, filepath.Join(manifestDir, entry.Name))
	}
	return hashFiles, nil
}

func cleanDuplicates(ref HashFiles, target HashFiles) HashFiles {
	slog.Info("cleaning duplicates", "target", target.NFiles(), "reference", ref.NFiles())
	for _, hash := range target.Hashes() {
		if !ref.Contains(hash) {
			slog.Debug("hash not found in reference", "hash", hash)
			continue
		}
		targetFiles := target.GetPaths(hash)
		refFiles := ref.GetPaths(hash)
		for _, duplicate := range getDuplicateNames(refFiles, targetFiles) {
			slog.Info("removing file", "hash", hash, "duplicate", duplicate)
		}
	}
	return target
}

func getDuplicateNames(ref []string, target []string) []string {
	duplicates := make([]string, 0, len(target))
	for _, targetFile := range target {
		for _, refFile := range ref {
			if filepath.Base(targetFile) == filepath.Base(refFile) {
				duplicates = append(duplicates, targetFile)
			}
		}
	}
	return duplicates
}
