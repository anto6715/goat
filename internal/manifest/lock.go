package manifest

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

func withExclusiveLock(root string, action func() error) error {
	lockPath := filepath.Join(root, LockFile)

	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return fmt.Errorf("failed to open lock file: %w", err)
	}
	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX); err != nil {
		_ = lockFile.Close()
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	defer func() {
		_ = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)
		_ = lockFile.Close()
	}()

	return action()
}
