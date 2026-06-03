//go:build !windows

package fsutil

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// mockFileInfo implements os.FileInfo
type mockFileInfo struct{}

func (m *mockFileInfo) Name() string       { return "mock" }
func (m *mockFileInfo) Size() int64        { return 0 }
func (m *mockFileInfo) Mode() fs.FileMode  { return 0 }
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) IsDir() bool        { return false }
func (m *mockFileInfo) Sys() interface{}   { return nil }

func TestGetInode(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "testinode")
	os.WriteFile(tmpFile, []byte("test"), 0644)

	info, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatalf("os.Stat() error = %v", err)
	}

	ino, ok := GetInode(tmpFile, info)
	if !ok {
		t.Error("GetInode() failed to retrieve inode")
	}
	if ino == 0 {
		t.Error("GetInode() returned inode 0")
	}

	// Test failure case
	mock := &mockFileInfo{}
	_, ok = GetInode("mock", mock)
	if ok {
		t.Error("GetInode() should have failed for mock file info")
	}
}
