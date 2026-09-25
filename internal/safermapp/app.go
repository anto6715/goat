package safermapp

import (
	"fmt"
	"log/slog"
	"os"
)

type Config struct {
	References []string
	Target     string
	MaxDepth   int
	Apply      bool
}

func Run(cfg Config) error {
	cfg, err := normalizeConfig(cfg)
	if err != nil {
		return err
	}

	slog.Info("Loading target files to be removed", "target", cfg.Target)
	targetIndex, err := loadIndex([]string{cfg.Target}, cfg.MaxDepth)
	if err != nil {
		return err
	}

	slog.Info("loading reference hash files", "reference", cfg.References)
	refIndex, err := loadIndex(cfg.References, cfg.MaxDepth)
	if err != nil {
		return err
	}
	plan := buildRemovalPlan(refIndex, targetIndex)

	if err := writeRemovalPlan(os.Stdout, plan); err != nil {
		return fmt.Errorf("failed to print removal plan: %w", err)
	}

	if len(plan) == 0 {
		return nil
	}

	if !cfg.Apply {
		fmt.Println("\nDry run: no files were removed")
		fmt.Println("Run again with --apply to remove verified duplicates")
		return nil
	}
	slog.Info("executing plan", "count", len(plan))
	return executePlan(plan)
}

// loadIndex builds an index from every manifest under roots.
func loadIndex(roots []string, depth int) (fileIndex, error) {
	index := make(fileIndex)

	for _, root := range roots {
		// get list of all available manifest files
		manifestFiles, err := loadManifest(root, depth)
		if err != nil {
			return nil, err
		}
		err = processManifests(index, manifestFiles)
		if err != nil {
			return nil, err
		}
	}
	return index, nil
}

func buildRemovalPlan(references fileIndex, targets fileIndex) []candidate {
	var plan []candidate

	for key, targetRecords := range targets {
		referenceRecords := references[key]
		if len(referenceRecords) == 0 {
			continue
		}

		for _, target := range targetRecords {
			plan = append(plan, candidate{
				references: referenceRecords,
				target:     target,
				key:        key,
			})
		}
	}
	return plan
}
