package main

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	"github.com/anto6715/goat/find"
	"github.com/anto6715/goat/internal/logging"
	"github.com/anto6715/goat/internal/tools"
)

type cli struct {
	Path    string `arg:"" name:"path" help:"Root directory to scan."`
	Workers int    `name:"workers" default:"2" help:"Number of workers."`
	Filter  string `name:"filter" default:"*" help:"Glob used to match file names."`
}

func main() {
	// Initialize logger
	logger := logging.New()
	slog.SetDefault(logger)

	// CLI args using kong
	var args cli
	kong.Parse(
		&args,
		kong.Name("gfind"),
		kong.Description("Find files under a root directory."),
		kong.UsageOnError(),
	)

	// Validate input arguments
	if err := tools.IsValidDir(args.Path); err != nil {
		slog.Error("invalid directory", "path", args.Path, "err", err)
		os.Exit(1)
	}

	// Abs calls also Clean
	root, err := filepath.Abs(args.Path)
	if err != nil {
		logger.Error("failed to get absolute path", "path", args.Path, "err", err)
		os.Exit(1)
	}

	// Execution
	if err := find.FindAndPrint(root, args.Filter); err != nil {
		logger.Error("find failed", "err", err)
		os.Exit(1)
	}
}
