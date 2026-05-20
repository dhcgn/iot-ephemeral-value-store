//go:build !windows

package storage

import (
	"fmt"

	"golang.org/x/sys/unix"
)

// diskFreeBytes returns the number of free bytes on the filesystem containing path.
func diskFreeBytes(path string) (int64, error) {
	var stat unix.Statfs_t
	if err := unix.Statfs(path, &stat); err != nil {
		return 0, fmt.Errorf("statfs failed: %w", err)
	}
	return int64(stat.Bavail) * int64(stat.Bsize), nil
}
