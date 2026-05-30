//go:build !windows

package fsutil

import (
	"os"
	"syscall"
)

// GetInode returns the inode number for a file on Unix systems.
func GetInode(_ string, info os.FileInfo) (uint64, bool) {
	if sys, ok := info.Sys().(*syscall.Stat_t); ok {
		return sys.Ino, true
	}
	return 0, false
}
