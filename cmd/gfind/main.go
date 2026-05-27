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
	Path     string `arg:"" name:"path" help:"Root directory to scan."`
	Filter   string `name:"filter" default:"*" help:"Glob used to match file names."`
	MaxDepth int    `name:"max-depth" default:"-1" help:"Maximum directory depth relative to root (-1 for unlimited)."`
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
	if args.MaxDepth < -1 {
		slog.Error("invalid max depth", "max_depth", args.MaxDepth, "err", "must be -1 or greater")
		os.Exit(1)
	}

	// Abs calls also Clean
	root, err := filepath.Abs(args.Path)
	if err != nil {
		logger.Error("failed to get absolute path", "path", args.Path, "err", err)
		os.Exit(1)
	}

	// Execution
	opts := find.DefaultOptions()
	opts.Filter = args.Filter
	opts.MaxDepth = args.MaxDepth
	if err := find.FindAndPrintWithOptions(root, opts); err != nil {
		logger.Error("find failed", "err", err)
		os.Exit(1)
	}
}
