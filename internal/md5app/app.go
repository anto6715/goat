package md5app

import (
	"fmt"
	"log/slog"
	"path/filepath"
	"sort"

	"github.com/anto6715/goat/find"
	"github.com/anto6715/goat/internal/manifest"
	"github.com/anto6715/goat/internal/tools"
)

type Options struct {
	Workers      int
	IgnoreErrors bool
	MaxDepth     int
	Filter       string
	Update       bool
}

func Run(root string, opts Options) error {
	// Safety Checks
	if err := tools.IsValidDir(root); err != nil {
		return fmt.Errorf("invalid directory %q: %w", root, err)
	}

	if opts.Workers < 1 {
		return fmt.Errorf("n-worker must be greater than 0")
	}

	// Find files under the root directory
	groups, err := findFiles(root, opts.MaxDepth, opts.Filter)
	if err != nil {
		return fmt.Errorf("failed to find files: %w", err)
	}

	// To keep processing order consistent
	for _, dir := range sortedDirectories(groups) {
		err := processDirectory(dir, groups, opts)
		if err != nil {
			return fmt.Errorf("failed to process directory %q: %w", dir, err)
		}
	}
	return nil
}

// Interact with the find package to find files under the root directory
func findFiles(root string, depth int, filter string) (map[string][]string, error) {
	findOpts := find.DefaultOptions()
	findOpts.MaxDepth = depth
	findOpts.Filter = filter
	groups, err := find.FindFilesWithDirs(root, findOpts)
	if err != nil {
		return nil, err
	}
	slog.Info("found directories", "count", len(groups))

	return groups, nil
}

func processDirectory(dir string, groups map[string][]string, opts Options) error {
	slog.Info("processing directory", "dir", dir)
	results, err := hashFiles(groups[dir], opts.Workers, opts.IgnoreErrors)
	if err != nil {
		return fmt.Errorf("failed to hash files in %q: %w", dir, err)
	}

	incoming := manifest.Manifest{
		Entries: buildManifestEntries(results),
	}
	err = writeManifest(dir, incoming, opts.Update)
	if err != nil {
		return fmt.Errorf("failed to write results: %w", err)
	}
	return nil
}

func writeManifest(dir string, incoming manifest.Manifest, update bool) error {
	// as default overwrite current manifest
	finalManifest := incoming

	// in this case finalManifest is a merge between the incoming manifest and the loaded current manifest
	if update {
		current, err := manifest.Load(dir)
		if err != nil {
			return fmt.Errorf("failed to load existing manifest for %q: %w", dir, err)
		}
		current.Merge(incoming)
		finalManifest = current
	}

	if err := manifest.Save(dir, finalManifest); err != nil {
		return fmt.Errorf("failed to save metadata for %q: %w", dir, err)
	}

	return nil
}

func sortedDirectories(groups map[string][]string) []string {
	dirs := make([]string, 0, len(groups))
	for dir := range groups {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)
	return dirs
}

func buildManifestEntries(hashResults []hashResult) map[string]manifest.Entry {
	entries := make(map[string]manifest.Entry, len(hashResults))
	for _, result := range hashResults {
		if result.err != nil {
			continue
		}

		filename := filepath.Base(result.path)
		entries[filename] = manifest.Entry{
			Name: filename,
			MD5:  result.sum,
		}
	}
	return entries
}
