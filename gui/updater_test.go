package gui

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractPathTraversal(t *testing.T) {
	tmpDir := t.TempDir()
	destPath := filepath.Join(tmpDir, "output")

	// Test ZIP traversal
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	f, _ := zw.Create("dupclean-../malicious.bin") // Zip path traversal attempt
	_, _ = f.Write([]byte("malicious"))
	zw.Close()

	zipPath := filepath.Join(tmpDir, "test.zip")
	_ = os.WriteFile(zipPath, buf.Bytes(), 0644)

	err := extractFromZip(zipPath, destPath)
	if err == nil {
		t.Error("Expected error for ZIP path traversal, got nil")
	}

	// Test TAR traversal
	bufTar := new(bytes.Buffer)
	gzw := gzip.NewWriter(bufTar)
	tw := tar.NewWriter(gzw)
	header := &tar.Header{
		Name: "dupclean-../../malicious.bin", // Tar path traversal attempt
		Size: 9,
		Mode: 0644,
	}
	tw.WriteHeader(header)
	tw.Write([]byte("malicious"))
	tw.Close()
	gzw.Close()

	tarPath := filepath.Join(tmpDir, "test.tar.gz")
	_ = os.WriteFile(tarPath, bufTar.Bytes(), 0644)

	err = extractFromTarGz(tarPath, destPath)
	if err == nil {
		t.Error("Expected error for TAR path traversal, got nil")
	}
}
