package md5app

import (
	"errors"
	"log/slog"
	"path/filepath"
	"sync"

	"github.com/anto6715/goat/filehash"
	"github.com/anto6715/goat/internal/manifest"
)

var errHashFailed = errors.New("failed to hash one or more files")

func hashFiles(paths []string, nWorker int, ignoreErrors bool) ([]hashResult, error) {
	// channel used by workers to receive jobs
	jobs := make(chan hashJob, nWorker)
	// channel used by workers to send results
	results := make(chan hashResult, nWorker)

	var workerWG sync.WaitGroup
	workerWG.Add(nWorker)

	for range nWorker {
		go func() {
			defer workerWG.Done()
			// Each worker stays alive until jobs is closed. Closing jobs is the
			// signal that no more files will arrive.
			for job := range jobs {
				sum, err := filehash.MD5Sum(job.path)
				results <- hashResult{
					path: job.path,
					sum:  sum,
					err:  err,
				}
			}
		}()
	}

	// The producer sends file paths to jobs, and workers send results back to the main goroutine.
	go func() {
		defer close(jobs)

		for _, path := range paths {
			if filepath.Base(path) == manifest.LegacyMetadataFile {
				continue
			}
			jobs <- hashJob{path: path}
		}
	}()

	go func() {
		// results must stay open until every worker has finished sending.
		// WaitGroup lets this goroutine close results exactly once, at the right time.
		workerWG.Wait()
		close(results)
	}()

	// gather results
	completed := make([]hashResult, 0, len(paths))
	failed := false
	for result := range results {
		completed = append(completed, result)
		slog.Info("hash completed", "hash", result.sum, "path", result.path)
		if result.err != nil {
			failed = true
			slog.Error("hash failed", "path", result.path, "err", result.err)
		}
	}

	if failed && !ignoreErrors {
		return nil, errHashFailed
	}
	return completed, nil
}
