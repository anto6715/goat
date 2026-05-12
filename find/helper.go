package find

import (
	"errors"
	"io/fs"
	"path/filepath"
)

func matchFile(pattern string, file string) (bool, error) {
	return filepath.Match(pattern, filepath.Base(file))
}

func shouldSkip(err error) bool {
	return errors.Is(err, fs.ErrPermission)
}
