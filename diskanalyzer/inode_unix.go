//go:build !windows

package diskanalyzer

import (
	"os"
	"syscall"
)

// getInode returns the inode number for a file on Unix systems.
func getInode(info os.FileInfo) (uint64, bool) {
	if sys, ok := info.Sys().(*syscall.Stat_t); ok {
		return sys.Ino, true
	}
	return 0, false
}
