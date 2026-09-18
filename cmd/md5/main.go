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
	index int
	path  string
}

type hashResult struct {
	index int
	path  string
	sum   string
	err   error
}

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

	if err := run(args, os.Stdout, os.Stderr); err != nil {
		slog.Error("md5 failed", "err", err)
		os.Exit(1)
	}
}

func run(args cli, stdout io.Writer, stderr io.Writer) error {
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
		if err := hashFiles(groups[dir], args.NWorker, stdout, stderr); err != nil {
			return fmt.Errorf("failed to hash files in %q: %w", dir, err)
		}
	}
	return nil
}

func hashFiles(paths []string, nWorker int, stdout io.Writer, stderr io.Writer) error {
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
					index: job.index,
					path:  job.path,
					sum:   sum,
					err:   err,
				}
			}
		}()
	}

	// The producer sends file paths to jobs, and workers send results back to the main goroutine.
	go func() {
		defer close(jobs)

		for index, path := range paths {
			jobs <- hashJob{index: index, path: path}
		}
	}()

	go func() {
		// results must stay open until every worker has finished sending.
		// WaitGroup lets this goroutine close results exactly once, at the right time.
		workerWG.Wait()
		close(results)
	}()

	// Workers finish at different times, so results can arrive out of order.
	// pending temporarily stores completed hashes until we have the next index
	// that should be printed.
	ordered := make(map[int]hashResult, len(paths))
	next := 0
	failed := false

	for result := range results {
		ordered[result.index] = result

		for {
			ready, ok := ordered[next]
			if !ok {
				break
			}

			if ready.err != nil {
				failed = true
				_, _ = fmt.Fprintf(stderr, "error: %s: %v\n", ready.path, ready.err)
			} else {
				_, _ = fmt.Fprintf(stdout, "%s %s\n", ready.sum, ready.path)
			}

			delete(ordered, next)
			next++
		}
	}

	if failed {
		return errHashFailed
	}
	return nil
}
