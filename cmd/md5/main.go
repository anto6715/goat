package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
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
	if err := tools.IsValidDir(args.Path); err != nil {
		return fmt.Errorf("invalid directory %q: %w", args.Path, err)
	}

	if args.NWorker < 1 {
		return fmt.Errorf("n-worker must be greater than 0")
	}

	return hashFiles(args.Path, args.NWorker, stdout, stderr)
}

func hashFiles(root string, nWorker int, stdout io.Writer, stderr io.Writer) error {
	// channel used by workers to receive jobs
	jobs := make(chan hashJob, nWorker)
	// channel used by workers to send results
	results := make(chan hashResult, nWorker)
	// channel used by the producer to send errors
	producerErrCh := make(chan error, 1)

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

		index := 0
		producerErrCh <- find.Find(root, "*", func(path string) error {
			jobs <- hashJob{index: index, path: path}
			index++
			return nil
		})
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
	pending := make(map[int]hashResult, nWorker)
	next := 0
	failed := false

	for result := range results {
		pending[result.index] = result

		for {
			ready, ok := pending[next]
			if !ok {
				break
			}

			if ready.err != nil {
				failed = true
				_, _ = fmt.Fprintf(stderr, "error: %s: %v\n", ready.path, ready.err)
			} else {
				_, _ = fmt.Fprintf(stdout, "%s %s\n", ready.sum, ready.path)
			}

			delete(pending, next)
			next++
		}
	}

	walkErr := <-producerErrCh
	switch {
	case walkErr != nil && failed:
		return errors.Join(
			fmt.Errorf("failed to find files under %q: %w", root, walkErr),
			errHashFailed,
		)
	case walkErr != nil:
		return fmt.Errorf("failed to find files under %q: %w", root, walkErr)
	case failed:
		return errHashFailed
	default:
		return nil
	}
}
