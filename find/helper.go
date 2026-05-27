package find

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func matchFile(pattern string, file string) (bool, error) {
	return filepath.Match(pattern, filepath.Base(file))
}

func shouldSkip(err error) bool {
	return errors.Is(err, fs.ErrPermission) || errors.Is(err, fs.ErrNotExist)
}

func dirDepth(path string) int {
	if path == "." || path == "" {
		return 0
	}

	return strings.Count(path, string(os.PathSeparator)) + 1
}

func fileDepth(path string) int {
	return dirDepth(filepath.Dir(path))
}
