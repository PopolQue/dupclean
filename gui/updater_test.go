package gui

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2/test"
)

// ...

func TestExtractFromTarGz(t *testing.T) {
	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "test.tar.gz")
	destPath := filepath.Join(tmpDir, "extracted")

	// Create a mock tar.gz
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gzw)

	content := []byte("binary content")
	header := &tar.Header{
		Name: "dupclean",
		Size: int64(len(content)),
		Mode: 0755,
	}
	tw.WriteHeader(header)
	tw.Write(content)
	tw.Close()
	gzw.Close()

	os.WriteFile(archivePath, buf.Bytes(), 0644)

	err := extractFromTarGz(archivePath, destPath)
	if err != nil {
		t.Fatalf("extractFromTarGz() error = %v", err)
	}

	got, _ := os.ReadFile(destPath)
	if string(got) != string(content) {
		t.Errorf("extracted content = %s, want %s", got, content)
	}

	// Test missing binary in archive
	buf.Reset()
	gzw = gzip.NewWriter(&buf)
	tw = tar.NewWriter(gzw)
	header = &tar.Header{
		Name: "otherfile",
		Size: 5,
	}
	tw.WriteHeader(header)
	tw.Write([]byte("other"))
	tw.Close()
	gzw.Close()
	os.WriteFile(archivePath, buf.Bytes(), 0644)

	err = extractFromTarGz(archivePath, destPath)
	if err == nil {
		t.Error("Expected error for archive without dupclean binary")
	}
}

func TestExtractFromZip(t *testing.T) {
	tmpDir := t.TempDir()
	archivePath := filepath.Join(tmpDir, "test.zip")
	destPath := filepath.Join(tmpDir, "extracted")

	// Create a mock zip
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	content := []byte("binary content")
	f, _ := zw.Create("dupclean.exe")
	f.Write(content)
	zw.Close()

	os.WriteFile(archivePath, buf.Bytes(), 0644)

	err := extractFromZip(archivePath, destPath)
	if err != nil {
		t.Fatalf("extractFromZip() error = %v", err)
	}

	got, _ := os.ReadFile(destPath)
	if string(got) != string(content) {
		t.Errorf("extracted content = %s, want %s", got, content)
	}

	// Test missing binary in archive
	buf.Reset()
	zw = zip.NewWriter(&buf)
	f, _ = zw.Create("otherfile")
	f.Write([]byte("other"))
	zw.Close()
	os.WriteFile(archivePath, buf.Bytes(), 0644)

	err = extractFromZip(archivePath, destPath)
	if err == nil {
		t.Error("Expected error for zip without dupclean binary")
	}
}

// ...

func TestVerifyHash(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "testfile")
	content := []byte("hello world")
	os.WriteFile(tmpFile, content, 0644)

	h := sha256.New()
	h.Write(content)
	expectedHash := hex.EncodeToString(h.Sum(nil))

	ok, err := verifyHash(tmpFile, expectedHash)
	if err != nil {
		t.Fatalf("verifyHash() error = %v", err)
	}
	if !ok {
		t.Error("verifyHash() should return true for correct hash")
	}

	ok, err = verifyHash(tmpFile, "wronghash")
	if err != nil {
		t.Fatalf("verifyHash() error = %v", err)
	}
	if ok {
		t.Error("verifyHash() should return false for incorrect hash")
	}

	// Test missing file
	_, err = verifyHash("/nonexistent/file", "hash")
	if err == nil {
		t.Error("Expected error for missing file")
	}
}

func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "src")
	dst := filepath.Join(tmpDir, "dst")
	content := []byte("copy test")

	os.WriteFile(src, content, 0644)

	err := copyFile(src, dst)
	if err != nil {
		t.Fatalf("copyFile() error = %v", err)
	}

	got, _ := os.ReadFile(dst)
	if string(got) != string(content) {
		t.Errorf("copyFile() content = %s, want %s", got, content)
	}

	// Test missing source
	err = copyFile("/nonexistent/file", dst)
	if err == nil {
		t.Error("Expected error for missing source file")
	}
}

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		current string
		latest  string
		want    bool
	}{
		{"v0.4.5", "v0.4.6", true},
		{"v0.4.6", "v0.4.5", false},
		{"v0.4.5", "v0.4.5", false},
		{"0.4.5", "0.4.6", true},
		{"v0.4.5.1", "v0.4.5.2", true},
		{"v0.4.5", "v0.5.0", true},
		{"invalid", "v0.4.6", false}, // Invalid current
		{"v0.4.5", "invalid", false}, // Invalid latest
	}

	for _, tc := range tests {
		got := isNewerVersion(tc.current, tc.latest)
		if got != tc.want {
			t.Errorf("isNewerVersion(%s, %s) = %v; want %v", tc.current, tc.latest, got, tc.want)
		}
	}
}

func TestIsValidUpdateURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"https://github.com/PopolQue/dupclean/releases", true},
		{"https://github.com/PopolQue/dupclean/archive/v0.4.9.0.tar.gz", true},
		{"https://raw.githubusercontent.com/PopolQue/dupclean/main/README.md", true},
		{"https://malicious.com", false},
		{":", false},
	}

	for _, tc := range tests {
		if got := isValidUpdateURL(tc.url); got != tc.expected {
			t.Errorf("isValidUpdateURL(%s) = %v; want %v", tc.url, got, tc.expected)
		}
	}
}

func TestNewUpdaterState(t *testing.T) {
	w := test.NewWindow(nil)
	state := NewUpdaterState(w)
	if state.Window != w {
		t.Error("Window not correctly set")
	}
}

func TestUpdaterWidget(t *testing.T) {
	_ = test.NewApp()
	w := test.NewWindow(nil)
	state := NewUpdaterState(w)

	// Just check it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("UpdaterWidget panicked: %v", r)
		}
	}()
	UpdaterWidget(state)
}

func TestShowUpdateDialog(t *testing.T) {
	_ = test.NewApp()
	w := test.NewWindow(nil)
	state := NewUpdaterState(w)
	release := &GitHubRelease{
		TagName: "v0.5.0",
		Body:    "## New features\n- something new\n## Installation\n- install guide",
	}

	// Just check it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("showUpdateDialog panicked: %v", r)
		}
	}()
	showUpdateDialog(state, release)
}

type mockClient struct {
	client *http.Client
}

func (m *mockClient) Get(url string) (*http.Response, error) {
	return m.client.Get(url)
}

func TestCheckForUpdates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		release := GitHubRelease{TagName: "v0.5.0", Body: "New features"}
		json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	// Inject mock
	oldClient := defaultClient
	oldAPI := githubAPI
	defer func() {
		defaultClient = oldClient
		githubAPI = oldAPI
	}()

	defaultClient = &mockClient{client: server.Client()}
	githubAPI = server.URL

	release, err := checkForUpdates()
	if err != nil {
		t.Fatalf("checkForUpdates() error = %v", err)
	}
	if release.TagName != "v0.5.0" {
		t.Errorf("Expected tag v0.5.0, got %s", release.TagName)
	}

	// Test error status
	server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	_, err = checkForUpdates()
	if err == nil {
		t.Error("Expected error for 404 status")
	}
}
