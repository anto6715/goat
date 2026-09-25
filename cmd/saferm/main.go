package main

import (
	"log/slog"
	"os"

	"github.com/alecthomas/kong"
	"github.com/anto6715/goat/internal/logging"
	"github.com/anto6715/goat/internal/safermapp"
)

type cli struct {
	References []string `name:"references" optional:"" help:"Directory to compute MD5 hashes for."`
	Target     string   `name:"target" optional:"" help:"Directory to remove files from."`
	MaxDepth   int      `short:"L" name:"max-depth" default:"-1" help:"Maximum directory depth relative to root (-1 for unlimited; 0 for root only)."`
	Apply      bool     `name:"apply" help:"Proceed with file removal based on the manifest."`
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

	err := safermapp.Run(safermapp.Config{
		References: args.References,
		Target:     args.Target,
		MaxDepth:   args.MaxDepth,
		Apply:      args.Apply,
	})

	if err != nil {
		slog.Error("saferm failed", "err", err)
		os.Exit(1)
	}
}
