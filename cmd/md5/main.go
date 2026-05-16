package main

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"sync"

	"github.com/alecthomas/kong"
	"github.com/anto6715/goat/filehash"
	"github.com/anto6715/goat/find"
	"github.com/anto6715/goat/internal/logging"
	"github.com/anto6715/goat/internal/tools"
)

type cli struct {
	Path string `arg:"" name:"path" help:"Directory to compute MD5 hash for."`
}

func main() {
	// Initialize logger
	logger := logging.New()
	slog.SetDefault(logger)

	// CLI args using kong
	var args cli
	kong.Parse(
		&args,
		kong.Name("MD5"),
		kong.Description("Compute MD5 hash of files under a root directory."),
		kong.UsageOnError(),
	)

	// Validate input arguments
	if err := tools.IsValidDir(args.Path); err != nil {
		slog.Error("invalid directory", "path", args.Path, "err", err)
		os.Exit(1)
	}

	// Load file paths from the directory
	files, err := find.FindFiles(args.Path, "*")
	if err != nil {
		slog.Error("failed to find files", "path", args.Path, "err", err)
		os.Exit(1)
	}

	// Set GOMAXPROCS to the number of CPU cores
	cores := runtime.NumCPU()
	runtime.GOMAXPROCS(cores)

	var wg sync.WaitGroup
	for _, file := range files {
		wg.Add(1)
		go func() {
			defer wg.Done()
			md5, err := filehash.MD5Sum(file)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			fmt.Println(md5, file)
		}()
	}

	wg.Wait()
}
