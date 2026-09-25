package safermapp

import (
	"fmt"
	"os"
	"path/filepath"
)

func normalizeConfig(cfg Config) (Config, error) {
	if cfg.MaxDepth < -1 {
		return Config{}, fmt.Errorf("max depth must be -1 or greater")
	}

	// at least 1 reference is required
	if len(cfg.References) == 0 {
		return Config{}, fmt.Errorf("at least one reference is required")
	}

	// not empty target
	if cfg.Target == "" {
		return Config{}, fmt.Errorf("target is required")
	}

	// validate and canonicalize target
	target, err := canonicalDir(cfg.Target)
	if err != nil {
		return Config{}, fmt.Errorf("invalid target: %w", err)
	}

	normalized := cfg
	normalized.Target = target
	normalized.References = make([]string, len(cfg.References))

	// validate and canonicalize references
	for i, reference := range cfg.References {
		normalizedReference, err := canonicalDir(reference)
		if err != nil {
			return Config{}, fmt.Errorf("invalid reference %d: %w", i, err)
		}
		normalized.References[i] = normalizedReference
	}

	// check if target is a reference
	for _, reference := range normalized.References {
		// reject if target is a reference
		if reference == target {
			return Config{}, fmt.Errorf("reference %q cannot be the target %q", reference, target)
		}

		// reject if target is a subdirectory of a reference
		overlap, err := pathsOverlap(reference, target)
		if err != nil {
			return Config{}, fmt.Errorf("compare reference %q and target %q: %w", reference, target, err)
		}
		if overlap {
			return Config{}, fmt.Errorf("reference %q and target %q cannot overlap", reference, target)
		}
	}

	return normalized, nil

}

func canonicalDir(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path is empty")
	}

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

func getManifestDir(manifestFile string) (string, error) {
	manifestDir, err := filepath.Abs(filepath.Dir(manifestFile))
	if err != nil {
		return "", err
	}
	return canonicalDir(manifestDir)
}

func pathsOverlap(reference, target string) (bool, error) {
	targetInsideReference, err := pathWithin(reference, target)
	if err != nil {
		return false, err
	}
	if targetInsideReference {
		return true, nil
	}

	referenceInsideTarget, err := pathWithin(target, reference)
	if err != nil {
		return false, fmt.Errorf("invalid path: %w", err)
	}
	if referenceInsideTarget {
		return true, nil
	}
	return false, nil
}

func pathWithin(parent, child string) (bool, error) {
	relative, err := filepath.Rel(parent, child)
	if err != nil {
		return false, err
	}
	return relative != "." && filepath.IsLocal(relative), nil
}
