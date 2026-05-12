package find

import (
	"io/fs"
	"path/filepath"
)

func Find(root string, nWorkers int, filter string, visit func(string) error) error {
	// A buffered channel here speeds up the performance
	fileChan := make(chan string, 100000)
	errChan := make(chan error, 1)

	// Walk the filesystem in parallel and collect results
	go func() {
		errChan <- walk(root, nWorkers, fileChan)
	}()

	// Collect here the worker output and apply visit
	for file := range fileChan {
		matched, err := matchFile(filter, file)
		if err != nil {
			return err
		}
		if matched {
			if err := visit(file); err != nil {
				return err
			}
		}
	}
	return <-errChan
}

func SerialFind(root string, filter string, visit func(string) error) error {
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if shouldSkip(err) {
				return nil
			}
			return err
		}
		// skip
		if d.IsDir() {
			return nil
		}
		matched, err := matchFile(filter, d.Name())
		if err != nil {
			return err
		}

		if !matched {
			return nil
		}

		if err := visit(path); err != nil {
			return err
		}
		return nil
	})

	return err
}
