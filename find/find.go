package find

import (
	"path/filepath"
)

func Find(root string, nWorkers int, filter string, visit func(string) error) error {
	// A buffered channel here speeds up the performance
	fileChan := make(chan string, 100000)

	// Walk the filesystem in parallel and collect results
	go walk(root, nWorkers, fileChan)

	// Collect here the worker output and apply visit
	for file := range fileChan {
		matched, err := filepath.Match(filter, filepath.Base(file))
		if err != nil {
			return err
		}
		if matched {
			if err := visit(file); err != nil {
				return err
			}
		}
	}
	return nil
}
