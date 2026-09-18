package md5app

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"

	"github.com/anto6715/goat/find"
	"github.com/anto6715/goat/internal/tools"
)

type Options struct {
	Workers      int
	IgnoreErrors bool
}

var errHashFailed = errors.New("failed to hash one or more files")

const legacyMetadataFile = ".dir_md5.txt"

func Run(root string, opts Options) error {
	// Safety Checks
	if err := tools.IsValidDir(root); err != nil {
		return fmt.Errorf("invalid directory %q: %w", root, err)
	}

	if opts.Workers < 1 {
		return fmt.Errorf("n-worker must be greater than 0")
	}

	// Find files under the root directory
	groups, err := find.FindFilesWithDirs(root, find.DefaultOptions())
	if err != nil {
		return fmt.Errorf("failed to find files: %w", err)
	}
	slog.Info("found directories", "count", len(groups))

	// To keep processing order consistent
	for _, dir := range sortedDirectories(groups) {
		slog.Info("processing directory", "dir", dir)

		hashResult, err := hashFiles(groups[dir], opts.Workers, opts.IgnoreErrors)
		if err != nil {
			return fmt.Errorf("failed to hash files in %q: %w", dir, err)
		}

		if err := saveMetadata(dir, hashResult); err != nil {
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

func saveMetadata(root string, hashes []hashResult) error {
	file, err := os.Create(root + "/" + legacyMetadataFile)
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}
	defer file.Close()

	for _, result := range hashes {
		filename := filepath.Base(result.path)
		_, err := fmt.Fprintln(file, result.sum, filename)
		if err != nil {
			return fmt.Errorf("failed to write metadata: %w", err)
		}
	}

	return nil
}
