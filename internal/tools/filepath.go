package tools

import (
	"errors"
	"os"
)

func IsValidDir(path string) error {
	// Safety Checks
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("path is not a directory")
	}
	return nil
}
