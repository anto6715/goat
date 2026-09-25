package safermapp

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
)

func executePlan(plan []candidate) error {
	for _, candidate := range plan {
		if err := validateCandidatePresence(candidate); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				slog.Warn("not exists", "path", candidate.target.path)
				continue
			}
			return fmt.Errorf("refuse to remove %q: %w", candidate.target.path, err)
		}

		if err := os.Remove(candidate.target.path); err != nil {
			return fmt.Errorf("remove %q: %w", candidate.target.path, err)
		}

		slog.Info("rm", "path", candidate.target.path)
	}

	return nil
}

func validateCandidatePresence(candidate candidate) error {
	targetInfo, err := os.Lstat(candidate.target.path)
	if err != nil {
		return fmt.Errorf("inspect target: %w", err)
	}
	if targetInfo.Mode().IsDir() {
		return fmt.Errorf("target is a directory")
	}

	for _, reference := range candidate.references {
		referenceInfo, err := os.Lstat(reference.path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("inspect reference %q: %w", reference.path, err)
		}
		if referenceInfo.Mode().IsRegular() {
			return nil
		}
	}

	return fmt.Errorf("no regular reference file exists")
}
