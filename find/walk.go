package find

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

func Walk(root string, nWorkers int, filter string) (*[]string, error) {
	// Initialize result
	result := make([]string, 0)
	
	// Communication management
	wg := sync.WaitGroup{}
	// A buffered channel is necessary to avoid deadlocks (otherwise workers immediately exit)
	dirChan := make(chan string, 100000)
	// A buffered channel here speeds up the performance
	fileChan := make(chan string, 100000)

	// Directory Scanner
	for range nWorkers {
		go dirScan(dirChan, fileChan, &wg)
	}

	// Start scan
	wg.Add(1)
	dirChan <- root

	// Wait in background the full scan
	go func() {
		wg.Wait()
		close(dirChan)
		close(fileChan)
	}()

	// Collect here the worker output
	for file := range fileChan {
		result = append(result, file)
		matched, err := filepath.Match(filter, filepath.Base(file))
		if err != nil {
			return nil, err
		}
		if matched {
			fmt.Println(file)
		}
	}

	return &result, nil
}

func dirScan(dirChan chan string, fileChan chan<- string, wg *sync.WaitGroup) {
	// Process directories as soon they are sent to the channel
	for dir := range dirChan {
		scanAndCollect(dir, dirChan, fileChan, wg)
	}
}

func scanAndCollect(path string, dirChan chan<- string, fileChan chan<- string, wg *sync.WaitGroup) {
	// Notify that a directory has been entirely scanned
	defer wg.Done()

	entries, err := os.ReadDir(path)
	// Permission error
	if err != nil {
		return
	}

	// iterate over directory content
	for _, entry := range entries {
		fullPath := filepath.Join(path, entry.Name())
		if entry.IsDir() {
			// Ask workers to scan subdirectories
			wg.Add(1)
			dirChan <- fullPath
		} else {
			fileChan <- fullPath
		}
	}
}
