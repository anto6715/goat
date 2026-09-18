package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sort"
	"sync"

	"github.com/alecthomas/kong"
	"github.com/anto6715/goat/filehash"
	"github.com/anto6715/goat/find"
	"github.com/anto6715/goat/internal/logging"
	"github.com/anto6715/goat/internal/tools"
)

type cli struct {
	Path    string `arg:"" name:"path" help:"Directory to compute MD5 hashes for."`
	NWorker int    `name:"workers" aliases:"nWorker" default:"2" help:"Number of hashing workers."`
}

type hashJob struct {
	path string
}

type hashResult struct {
	path string
	sum  string
	err  error
}

const legacyMetadataFile = ".dir_md5.txt"

var errHashFailed = errors.New("failed to hash one or more files")

func main() {
	logger := logging.New()
	slog.SetDefault(logger)

	// CLI args using kong
	var args cli
	kong.Parse(
		&args,
		kong.Name("md5"),
		kong.Description("Compute MD5 hashes of files under a root directory."),
		kong.UsageOnError(),
	)

	if err := run(args, os.Stdout); err != nil {
		slog.Error("md5 failed", "err", err)
		os.Exit(1)
	}
}

func run(args cli, stdout io.Writer) error {
	// Safety Checks
	if err := tools.IsValidDir(args.Path); err != nil {
		return fmt.Errorf("invalid directory %q: %w", args.Path, err)
	}

	if args.NWorker < 1 {
		return fmt.Errorf("n-worker must be greater than 0")
	}

	// Find files under the root directory
	groups, err := find.FindFilesWithDirs(args.Path, find.DefaultOptions())
	if err != nil {
		return fmt.Errorf("failed to find files: %w", err)
	}

	dirs := make([]string, 0, len(groups))
	for dir := range groups {
		dirs = append(dirs, dir)
	}
	slog.Info("found directories", "count", len(dirs))

	// To keep processing order consistent
	sort.Strings(dirs)
	for _, dir := range dirs {
		slog.Info("processing directory", "dir", dir)
		hashResult, err := hashFiles(groups[dir], args.NWorker)
		if err != nil {
			return fmt.Errorf("failed to hash files in %q: %w", dir, err)
		}

		if err := saveMetadata(dir, hashResult); err != nil {
			return fmt.Errorf("failed to save metadata for %q: %w", dir, err)
		}
	}
	return nil
}

func hashFiles(paths []string, nWorker int) ([]hashResult, error) {
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
	completed := make([]hashResult, len(paths))
	failed := false
	for result := range results {
		completed = append(completed, result)

		if result.err != nil {
			failed = true
			slog.Error("hash failed", "path", result.path, "err", result.err)
		}
	}

	if failed {
		return nil, errHashFailed
	}
	return completed, nil
}

func saveMetadata(root string, hashes []hashResult) error {
	file, err := os.Create(root + "/" + legacyMetadataFile)
	if err != nil {
		return fmt.Errorf("failed to create metadata file: %w", err)
	}
	defer file.Close()

	for _, result := range hashes {
		_, err := fmt.Fprintln(file, result.path, result.sum)
		if err != nil {
			return fmt.Errorf("failed to write metadata: %w", err)
		}
	}

	return nil
}
