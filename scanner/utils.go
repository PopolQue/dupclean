package scanner

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// MemoryWarningThreshold is the number of files at which we warn about memory usage
const MemoryWarningThreshold = 100000

// DefaultPartialHashSize is the default number of bytes to hash for partial file hashing.
// 8KB provides a good balance between speed and collision avoidance for initial filtering.
// Files with different partial hashes are guaranteed different; files with matching
// partial hashes proceed to full hash verification.
const DefaultPartialHashSize = 8 * 1024 // 8KB

// DefaultComparisonBufferSize is the default buffer size for byte-by-byte file comparison.
// 32KB buffers provide good throughput for sequential file I/O while keeping memory
// usage reasonable. Used in filesIdentical() for final duplicate verification.
const DefaultComparisonBufferSize = 32 * 1024 // 32KB

// partialHashSize is the legacy constant for backwards compatibility.
//
// Deprecated: Use DefaultPartialHashSize instead.
const partialHashSize = DefaultPartialHashSize

// hashFilePartial computes SHA256 of the first N bytes of a file
func hashFilePartial(path string, size int64) (string, error) {
	// #nosec G304
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	lr := io.LimitReader(f, size)
	if _, err := io.Copy(h, lr); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// hashFileFull computes SHA256 of the entire file
func hashFileFull(path string) (string, os.FileInfo, error) {
	// #nosec G304
	f, err := os.Open(path)
	if err != nil {
		return "", nil, err
	}
	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil {
		return "", nil, err
	}

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", nil, err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), info, nil
}

// filesIdentical performs byte-by-byte comparison of two files
func filesIdentical(path1, path2 string) (bool, error) {
	// #nosec G304
	f1, err := os.Open(path1)
	if err != nil {
		return false, err
	}
	defer func() { _ = f1.Close() }()

	// #nosec G304
	f2, err := os.Open(path2)
	if err != nil {
		return false, err
	}
	defer func() { _ = f2.Close() }()

	// Use configurable buffer size for comparison
	buf1 := make([]byte, DefaultComparisonBufferSize)
	buf2 := make([]byte, DefaultComparisonBufferSize)

	for {
		n1, err1 := f1.Read(buf1)
		n2, err2 := f2.Read(buf2)

		if n1 != n2 {
			return false, nil
		}

		if !bytes.Equal(buf1[:n1], buf2[:n2]) {
			return false, nil
		}

		if err1 == io.EOF || err2 == io.EOF {
			return err1 == io.EOF && err2 == io.EOF, nil
		}

		if err1 != nil && err1 != io.EOF {
			return false, err1
		}
		if err2 != nil && err2 != io.EOF {
			return false, err2
		}
	}
}
