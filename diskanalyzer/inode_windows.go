//go:build windows

package diskanalyzer

import (
	"os"
	"syscall"
)

// getInode returns the file index (inode equivalent) for a file on Windows.
// This allows the analyzer to skip hard links on NTFS, matching Unix behavior.
func getInode(path string, _ os.FileInfo) (uint64, bool) {
	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, false
	}

	// Open handle with minimal permissions (0 access) to get metadata
	// FILE_SHARE_READ|FILE_SHARE_WRITE|FILE_SHARE_DELETE allows others to still access the file
	// FILE_FLAG_BACKUP_SEMANTICS allows opening directories if needed
	h, err := syscall.CreateFile(
		pathPtr,
		0,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
	if err != nil {
		return 0, false
	}
	defer syscall.CloseHandle(h)

	var info syscall.ByHandleFileInformation
	err = syscall.GetFileInformationByHandle(h, &info)
	if err != nil {
		return 0, false
	}

	// File index is the NTFS equivalent of an inode
	inode := uint64(info.FileIndexHigh)<<32 | uint64(info.FileIndexLow)
	return inode, true
}
