package safermapp

import (
	"fmt"
	"os"
	"path/filepath"
)

func validateConfig(cfg Config) error {
	if cfg.MaxDepth < -1 {
		return fmt.Errorf("max depth must be > -1")
	}

	// at least 1 reference is required
	if len(cfg.References) == 0 {
		return fmt.Errorf("at least one reference is required")
	}

	// not empty target
	if cfg.Target == "" {
		return fmt.Errorf("target is required")
	}

	// validate and canonicalize target
	target, err := canonicalDir(cfg.Target)
	if err != nil {
		return err
	}
	cfg.Target = target

	// validate and canonicalize references
	for i, reference := range cfg.References {
		ref, err := canonicalDir(reference)
		if err != nil {
			return err
		}
		cfg.References[i] = ref
	}

	// check if target is a reference
	for _, reference := range cfg.References {
		if reference == target {
			return fmt.Errorf("reference '%s' cannot be the target '%s'", reference, target)
		}
	}

	return nil

}

func canonicalDir(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(resolved)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%q is not a directory", path)
	}

	return resolved, nil
}
