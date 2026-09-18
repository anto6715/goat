package main

import (
	"log/slog"
	"os"

	"github.com/alecthomas/kong"
	"github.com/anto6715/goat/internal/logging"
	"github.com/anto6715/goat/internal/md5app"
)

type cli struct {
	Path         string `arg:"" name:"path" help:"Directory to compute MD5 hashes for."`
	NWorker      int    `name:"workers" aliases:"nWorker" default:"2" help:"Number of hashing workers."`
	IgnoreErrors bool   `name:"ignore-errors" default:"false" help:"Ignore errors and continue processing."`
}


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

	err := md5app.Run(args.Path, md5app.Options{
		Workers:      args.NWorker,
		IgnoreErrors: args.IgnoreErrors,
	})
	
	if err != nil {
		slog.Error("md5 failed", "err", err)
		os.Exit(1)
	}
}
