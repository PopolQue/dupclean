package scanner

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

const partialHashSize = 8 * 1024 // 8KB

// hashFilePartial computes SHA256 of the first N bytes of a file
func hashFilePartial(path string, size int64) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	lr := io.LimitReader(f, size)
	if _, err := io.Copy(h, lr); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// hashFileFull computes SHA256 of the entire file
func hashFileFull(path string) (string, os.FileInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", nil, err
	}
	defer f.Close()

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
	f1, err := os.Open(path1)
	if err != nil {
		return false, err
	}
	defer f1.Close()

	f2, err := os.Open(path2)
	if err != nil {
		return false, err
	}
	defer f2.Close()

	const bufSize = 32 * 1024 // 32KB buffers
	buf1 := make([]byte, bufSize)
	buf2 := make([]byte, bufSize)

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
