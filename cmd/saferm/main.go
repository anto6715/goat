package main

import (
	"log/slog"
	"os"

	"github.com/alecthomas/kong"
	"github.com/anto6715/goat/internal/logging"
	"github.com/anto6715/goat/internal/safermapp"
)

type cli struct {
	Path     string `arg:"" name:"path" help:"Directory to compute MD5 hashes for."`
	MaxDepth int    `short:"L" name:"max-depth" default:"-1" help:"Maximum directory depth relative to root (-1 for unlimited; 0 for root only)."`
}

func main() {
	logger := logging.New()
	slog.SetDefault(logger)

	// CLI args using kong
	var args cli
	kong.Parse(
		&args,
		kong.Name("saferm"),
		kong.Description("Safe remove files using manifest"),
		kong.UsageOnError(),
	)

	err := safermapp.Run(args.Path, safermapp.Options{
		Depth: args.MaxDepth,
	})

	if err != nil {
		slog.Error("saferm failed", "err", err)
		os.Exit(1)
	}
}
