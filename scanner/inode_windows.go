//go:build windows

package scanner

import "os"

// getInode returns a dummy inode on Windows (NTFS hard links are rare).
func getInode(_ os.FileInfo) (uint64, bool) {
	return 0, false
}
