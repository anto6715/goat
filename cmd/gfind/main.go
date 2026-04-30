package main

import (
	"flag"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/anto6715/goat/find"
	"github.com/anto6715/goat/internal/logging"
)

func main() {
	logger := logging.New()
	slog.SetDefault(logger)
	
	// User args
	userPath := flag.String("path", "", "root directory")
	nWorkers := flag.Int("workers", 1, "number of workers")
	filter := flag.String("filter", "*", "filter")
	flag.Parse()

	// Safety Checks
	if *userPath == "" {
		logger.Error("usage: go run main.go [-path path]")
		os.Exit(1)
	}

	if _, err := os.Stat(*userPath); os.IsNotExist(err) {
		logger.Error("Not exists", "path", *userPath)
		os.Exit(1)
	}

	// Abs calls also Clean
	root, err := filepath.Abs(*userPath)
	if err != nil {
		logger.Error("Error getting absolute path", "err", err)
		os.Exit(1)
	}

	// Execution
	//start := time.Now()
	find.Walk(root, *nWorkers, *filter)
}
