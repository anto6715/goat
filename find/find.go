package find

import (
	"fmt"
	"io/fs"
	"path/filepath"
)

// Find is the single-threaded reference implementation of Find.
func Find(root string, filter string, visit func(string) error) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if shouldSkip(err) {
				return nil
			}
			return err
		}

		if d.IsDir() {
			return nil
		}

		matched, err := matchFile(filter, path)
		if err != nil {
			return err
		}

		if !matched {
			return nil
		}

		return visit(path)
	})
}

func FindAndPrint(root string, filter string) error {
	return Find(root, filter, func(path string) error {
		_, err := fmt.Println(path)
		return err
	})
}
