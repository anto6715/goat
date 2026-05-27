package find

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Options struct {
	Filter   string
	MaxDepth int // -1 means no limit
}

const (
	defaultFilter   = "*"
	defaultMaxDepth = -1
)

func DefaultOptions() Options {
	return Options{
		Filter:   defaultFilter,
		MaxDepth: defaultMaxDepth,
	}
}

// FindWithOptions walks root and calls visit for every matching file.
func FindWithOptions(root string, opts Options, visit func(string) error) error {
	root = filepath.Clean(root)

	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if shouldSkip(err) {
				return nil
			}
			return err
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		if d.IsDir() {
			if opts.MaxDepth >= 0 && dirDepth(rel) > opts.MaxDepth {
				// Returning SkipDir stops WalkDir from descending further into
				// directories that are already deeper than the allowed limit.
				return fs.SkipDir
			}
			return nil
		}

		if opts.MaxDepth >= 0 && fileDepth(rel) > opts.MaxDepth {
			return nil
		}

		matched, err := matchFile(opts.Filter, rel)
		if err != nil {
			return err
		}

		if !matched {
			return nil
		}

		return visit(path)
	})
}

func FindAndPrintWithOptions(root string, opts Options) error {
	return FindWithOptions(root, opts, func(path string) error {
		_, err := fmt.Println(path)
		return err
	})
}

func FindFilesWithOptions(root string, opts Options) ([]string, error) {
	var paths []string
	err := FindWithOptions(root, opts, func(path string) error {
		paths = append(paths, path)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return paths, nil
}
