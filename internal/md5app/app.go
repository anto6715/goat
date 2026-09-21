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
	findOpts := find.DefaultOptions()
	findOpts.MaxDepth = opts.MaxDepth
	findOpts.Filter = opts.Filter
	groups, err := find.FindFilesWithDirs(root, findOpts)
	if err != nil {
		return fmt.Errorf("failed to find files: %w", err)
	}
	slog.Info("found directories", "count", len(groups))

	// To keep processing order consistent
	for _, dir := range sortedDirectories(groups) {
		slog.Info("processing directory", "dir", dir)

		results, err := hashFiles(groups[dir], opts.Workers, opts.IgnoreErrors)
		if err != nil {
			return fmt.Errorf("failed to hash files in %q: %w", dir, err)
		}

		m := manifest.Manifest{
			Entries: buildManifestEntries(results),
		}

		if err := manifest.Save(dir, m); err != nil {
			return fmt.Errorf("failed to save metadata for %q: %w", dir, err)
		}
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
