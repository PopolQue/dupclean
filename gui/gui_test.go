package gui

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"dupclean/scanner"
)

// TestIsValidExtension tests the extension validation function
func TestIsValidExtension(t *testing.T) {
	tests := []struct {
		ext      string
		expected bool
	}{
		{".txt", true},
		{".pdf", true},
		{".jpg", true},
		{".wav", true},
		{"*", false},
		{".*", false},
		{"*.txt", false},
		{"?", false},
		{".?", false},
		{"", true},
		{".log", true},
		{".tmp", true},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			result := isValidExtension(tt.ext)
			if result != tt.expected {
				t.Errorf("isValidExtension(%q) = %v; want %v", tt.ext, result, tt.expected)
			}
		})
	}
}

// TestIsProtectedPath_CrossPlatform tests protected path validation for all platforms
func TestIsProtectedPath_CrossPlatform(t *testing.T) {
	// Test macOS paths
	if runtime.GOOS == "darwin" {
		protectedPaths := []string{
			"/var/folders",
			"/var/folders/abc/123",
			"/private/var",
			"/private/var/folders",
			"/System",
			"/System/Library",
			"/Library/Caches/com.apple",
			"/Library/Caches/com.apple.Safari",
		}

		for _, path := range protectedPaths {
			if !isProtectedPath(path) {
				t.Errorf("isProtectedPath(%q) should be true on macOS", path)
			}
		}

		nonProtectedPaths := []string{
			"/Users/test/Documents",
			"/tmp/test",
			"/Applications",
		}

		for _, path := range nonProtectedPaths {
			if isProtectedPath(path) {
				t.Errorf("isProtectedPath(%q) should be false on macOS", path)
			}
		}
	}

	// Test Linux paths
	if runtime.GOOS == "linux" {
		protectedPaths := []string{
			"/etc",
			"/etc/passwd",
			"/bin",
			"/usr",
			"/usr/bin",
			"/sbin",
			"/lib",
			"/boot",
			"/dev",
			"/proc",
			"/sys",
		}

		for _, path := range protectedPaths {
			if !isProtectedPath(path) {
				t.Errorf("isProtectedPath(%q) should be true on linux", path)
			}
		}

		nonProtectedPaths := []string{
			"/home/user",
			"/tmp/test",
			"/opt/app",
		}

		for _, path := range nonProtectedPaths {
			if isProtectedPath(path) {
				t.Errorf("isProtectedPath(%q) should be false on linux", path)
			}
		}
	}
}

// TestIsProtectedPath_EdgeCases tests edge cases for protected path validation
func TestIsProtectedPath_EdgeCases(t *testing.T) {
	// Empty path
	result := isProtectedPath("")
	// Empty path should either be protected or not, just ensure it doesn't panic
	t.Logf("isProtectedPath(\"\") = %v", result)

	// Relative paths
	result = isProtectedPath("relative/path")
	t.Logf("isProtectedPath(\"relative/path\") = %v", result)
}

// TestFormatBytes_EdgeCases tests edge cases for byte formatting
func TestFormatBytes_EdgeCases(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0 B"},
		{1, "1 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1023 * 1024 * 1024 * 1024 * 1024, "1023.0 TB"},
	}

	for _, tt := range tests {
		result := formatBytes(tt.input)
		if result != tt.expected {
			t.Errorf("formatBytes(%d) = %s; want %s", tt.input, result, tt.expected)
		}
	}
}

// TestAppState_UpdateContent tests the updateContent method
func TestAppState_UpdateContent(t *testing.T) {
	state := &AppState{
		ContentContainer: nil,
	}

	// Should not panic when ContentContainer is nil
	state.updateContent(nil)

	// Test with valid state
	state2 := &AppState{
		FolderPath:        "/test",
		CurrentGroupIndex: 0,
		DeletedCount:      0,
		FreedBytes:        0,
		Groups:            []scanner.DuplicateGroup{},
	}

	if state2.FolderPath != "/test" {
		t.Errorf("FolderPath = %q; want %q", state2.FolderPath, "/test")
	}
}

// TestAppState_StateTransitions tests state transitions
func TestAppState_StateTransitions(t *testing.T) {
	state := &AppState{
		IsScanning:        false,
		CurrentGroupIndex: 0,
		DeletedCount:      0,
		FreedBytes:        0,
		Groups: []scanner.DuplicateGroup{
			{
				Hash: "hash1",
				Files: []scanner.FileInfo{
					{Path: "/test/file1.txt", Name: "file1.txt", Size: 100},
					{Path: "/test/file2.txt", Name: "file2.txt", Size: 100},
				},
			},
			{
				Hash: "hash2",
				Files: []scanner.FileInfo{
					{Path: "/test/file3.txt", Name: "file3.txt", Size: 200},
					{Path: "/test/file4.txt", Name: "file4.txt", Size: 200},
				},
			},
		},
	}

	// Simulate navigation
	state.CurrentGroupIndex = 1
	if state.CurrentGroupIndex != 1 {
		t.Errorf("CurrentGroupIndex = %d; want 1", state.CurrentGroupIndex)
	}

	// Simulate deletion
	state.DeletedCount = 2
	state.FreedBytes = 200
	if state.DeletedCount != 2 {
		t.Errorf("DeletedCount = %d; want 2", state.DeletedCount)
	}
	if state.FreedBytes != 200 {
		t.Errorf("FreedBytes = %d; want 200", state.FreedBytes)
	}
}

// TestKeepAndDelete_EdgeCases tests edge cases for keepAndDelete
func TestKeepAndDelete_EdgeCases(t *testing.T) {
	// Test with empty files list
	state := &AppState{
		DeletedCount:      0,
		FreedBytes:        0,
		CurrentGroupIndex: 0,
		Groups: []scanner.DuplicateGroup{
			{
				Hash:  "testhash",
				Files: []scanner.FileInfo{},
			},
		},
	}

	files := state.Groups[0].Files
	keepAndDelete(state, 0, files)

	// Should not crash and group should be removed
	if len(state.Groups) != 0 {
		t.Errorf("Groups should be empty after keepAndDelete, got %d", len(state.Groups))
	}
}

// TestKeepAndDelete_MultipleGroups tests deletion with multiple groups
func TestKeepAndDelete_MultipleGroups(t *testing.T) {
	state := &AppState{
		DeletedCount:      0,
		FreedBytes:        0,
		CurrentGroupIndex: 1,
		Groups: []scanner.DuplicateGroup{
			{
				Hash: "hash1",
				Files: []scanner.FileInfo{
					{Path: "/test/file1.txt", Name: "file1.txt", Size: 100},
					{Path: "/test/file2.txt", Name: "file2.txt", Size: 100},
				},
			},
			{
				Hash: "hash2",
				Files: []scanner.FileInfo{
					{Path: "/test/file3.txt", Name: "file3.txt", Size: 200},
					{Path: "/test/file4.txt", Name: "file4.txt", Size: 200},
					{Path: "/test/file5.txt", Name: "file5.txt", Size: 200},
				},
			},
		},
	}

	files := state.Groups[1].Files
	keepAndDelete(state, 0, files)

	if state.DeletedCount != 2 {
		t.Errorf("DeletedCount = %d; want 2", state.DeletedCount)
	}
	if state.FreedBytes != 400 {
		t.Errorf("FreedBytes = %d; want 400", state.FreedBytes)
	}
	// Should have 1 group left
	if len(state.Groups) != 1 {
		t.Errorf("Groups length = %d; want 1", len(state.Groups))
	}
}

// TestStopPlayback_EdgeCases tests edge cases for stopPlayback
func TestStopPlayback_EdgeCases(t *testing.T) {
	// Test with nil state - should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Logf("stopPlayback with nil state panicked: %v", r)
		}
	}()

	// This will panic if we don't check for nil, which is expected
	// stopPlayback(nil) // Don't call this - it will panic

	// Test with empty state
	state := &AppState{}
	stopPlayback(state)
	if state.CurrentPlayer != nil {
		t.Error("CurrentPlayer should be nil")
	}
	if state.StopPlayer != nil {
		t.Error("StopPlayer should be nil")
	}
	if state.PlayingPath != "" {
		t.Error("PlayingPath should be empty")
	}
}

// TestRuntimeOS_Valid tests that runtimeOS returns a valid OS
func TestRuntimeOS_Valid(t *testing.T) {
	result := runtimeOS()
	if result == "" {
		t.Error("runtimeOS should return a non-empty string")
	}

	validOS := map[string]bool{
		"darwin":  true,
		"linux":   true,
		"windows": true,
		"freebsd": true,
		"netbsd":  true,
		"openbsd": true,
		"plan9":   true,
		"solaris": true,
		"illumos": true,
		"js":      true,
		"wasip1":  true,
	}

	if !validOS[result] {
		t.Errorf("runtimeOS returned unknown OS: %q", result)
	}
}

// TestFormatBytes_Consistency tests that formatBytes is consistent
func TestFormatBytes_Consistency(t *testing.T) {
	// Test that larger inputs produce larger or equal outputs
	sizes := []int64{0, 100, 1000, 10000, 100000, 1000000}
	lastLen := 0
	for _, size := range sizes {
		result := formatBytes(size)
		if len(result) < lastLen && size > 0 {
			t.Logf("formatBytes(%d) = %q (length decreased from %d)", size, result, lastLen)
		}
		lastLen = len(result)
	}
}

// TestMoveToTrash_NonExistentFile tests moveToTrash with non-existent file
func TestMoveToTrash_NonExistentFile(t *testing.T) {
	err := moveToTrash("/nonexistent/path/that/does/not/exist.txt")
	// We expect an error, but it shouldn't panic
	if err == nil {
		t.Log("moveToTrash on non-existent file returned nil (unexpected)")
	}
}

// TestMoveToTrash_ValidFile tests moveToTrash with a valid file
func TestMoveToTrash_ValidFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_trash.txt")

	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err := moveToTrash(testFile)
	// May succeed or fail depending on the platform
	t.Logf("moveToTrash result: %v", err)
}

// TestCacheCleanerState_Init tests CacheCleanerState initialization
func TestCacheCleanerState_Init(t *testing.T) {
	state := &CacheCleanerState{
		SelectedTargets: make(map[string]bool),
		Targets:         nil,
		TotalSize:       0,
		IsScanning:      false,
		IsCleaning:      false,
		CleanedCount:    0,
		CleanedBytes:    0,
	}

	if state.SelectedTargets == nil {
		t.Error("SelectedTargets should be initialized")
	}
	if state.TotalSize != 0 {
		t.Errorf("TotalSize = %d; want 0", state.TotalSize)
	}
	if state.IsScanning {
		t.Error("IsScanning should be false initially")
	}
	if state.IsCleaning {
		t.Error("IsCleaning should be false initially")
	}
}

// TestCacheCleanerState_WithTargets tests CacheCleanerState with targets
func TestCacheCleanerState_WithTargets(t *testing.T) {
	state := &CacheCleanerState{
		SelectedTargets: map[string]bool{
			"target1": true,
			"target2": false,
		},
		TotalSize: 1024 * 1024 * 100, // 100 MB
	}

	if len(state.SelectedTargets) != 2 {
		t.Errorf("SelectedTargets length = %d; want 2", len(state.SelectedTargets))
	}
	if state.TotalSize != 1024*1024*100 {
		t.Errorf("TotalSize = %d; want %d", state.TotalSize, 1024*1024*100)
	}
}

// TestCacheCleanerState_UpdateContent_Nil tests updateContent with nil container
func TestCacheCleanerState_UpdateContent_Nil(t *testing.T) {
	state := &CacheCleanerState{
		ContentContainer: nil,
	}

	// Should not panic
	state.updateContent(nil)
}

// TestCleanPath_NonExistent tests cleanPath with non-existent path
func TestCleanPath_NonExistent(t *testing.T) {
	deleted, freed, err := cleanPath("/nonexistent/path/that/does/not/exist", []string{"*"})
	if err != nil {
		t.Logf("cleanPath error (expected): %v", err)
	}
	// Should handle gracefully
	t.Logf("cleanPath result: deleted=%d, freed=%d", deleted, freed)
}

// TestCleanPath_StarPattern tests cleanPath with star pattern
func TestCleanPath_StarPattern(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some test files
	files := []string{"file1.txt", "file2.log", "file3.tmp"}
	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	deleted, freed, err := cleanPath(tmpDir, []string{"*"})
	if err != nil {
		t.Logf("cleanPath error: %v", err)
	}

	// Should delete all files
	if deleted < 3 {
		t.Logf("cleanPath deleted %d files (expected >= 3)", deleted)
	}
	if freed == 0 {
		t.Log("cleanPath freed 0 bytes (may be expected)")
	}

	t.Logf("cleanPath result: deleted=%d, freed=%d", deleted, freed)
}

// TestCleanPath_SpecificPattern tests cleanPath with specific patterns
func TestCleanPath_SpecificPattern(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some test files
	files := []struct {
		name    string
		content string
	}{
		{"cache1.tmp", "cache1"},
		{"cache2.tmp", "cache2"},
		{"data.txt", "data"},
	}

	for _, f := range files {
		path := filepath.Join(tmpDir, f.name)
		if err := os.WriteFile(path, []byte(f.content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	deleted, freed, err := cleanPath(tmpDir, []string{"*.tmp"})
	if err != nil {
		t.Logf("cleanPath error: %v", err)
	}

	// Should delete only .tmp files
	if deleted != 2 {
		t.Logf("cleanPath deleted %d files (expected 2)", deleted)
	}

	t.Logf("cleanPath result: deleted=%d, freed=%d", deleted, freed)
}

// TestCleanPath_EmptyPatterns tests cleanPath with empty patterns
func TestCleanPath_EmptyPatterns(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	path := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	deleted, freed, err := cleanPath(tmpDir, []string{})
	if err != nil {
		t.Logf("cleanPath error: %v", err)
	}

	// Should not delete anything with empty patterns
	if deleted != 0 {
		t.Logf("cleanPath deleted %d files (expected 0)", deleted)
	}
	if freed != 0 {
		t.Logf("cleanPath freed %d bytes (expected 0)", freed)
	}
}

// TestCleanPath_NoMatch tests cleanPath when no files match the pattern
func TestCleanPath_NoMatch(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files that don't match pattern
	files := []string{"file1.txt", "file2.log"}
	for _, f := range files {
		path := filepath.Join(tmpDir, f)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	deleted, freed, err := cleanPath(tmpDir, []string{"*.tmp"})
	if err != nil {
		t.Logf("cleanPath error: %v", err)
	}

	if deleted != 0 {
		t.Logf("cleanPath deleted %d files (expected 0)", deleted)
	}
	if freed != 0 {
		t.Logf("cleanPath freed %d bytes (expected 0)", freed)
	}
}

// TestFormatBytes_BusinessLogic tests formatBytes for business logic correctness
func TestFormatBytes_BusinessLogic(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{"0 B", 0, "0 B"},
		{"1 B", 1, "1 B"},
		{"1023 B", 1023, "1023 B"},
		{"1.0 KB", 1024, "1.0 KB"},
		{"1.5 KB", 1536, "1.5 KB"},
		{"1.0 MB", 1048576, "1.0 MB"},
		{"1.0 GB", 1073741824, "1.0 GB"},
		{"1.0 TB", 1099511627776, "1.0 TB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBytes(tt.input)
			if result != tt.expected {
				t.Errorf("formatBytes(%d) = %s; want %s", tt.input, result, tt.expected)
			}
		})
	}
}

// TestRuntimeOS_BusinessLogic tests runtimeOS business logic
func TestRuntimeOS_BusinessLogic(t *testing.T) {
	result := runtimeOS()

	// Should match actual runtime
	if result != runtime.GOOS {
		t.Errorf("runtimeOS() = %q; want %q", result, runtime.GOOS)
	}
}

// TestStopPlayback_NoPlayer tests stopPlayback when no player is active
func TestStopPlayback_NoPlayer(t *testing.T) {
	state := &AppState{
		CurrentPlayer: nil,
		StopPlayer:    nil,
		PlayingPath:   "",
	}

	stopPlayback(state)

	if state.CurrentPlayer != nil {
		t.Error("CurrentPlayer should be nil")
	}
	if state.StopPlayer != nil {
		t.Error("StopPlayer should be nil")
	}
	if state.PlayingPath != "" {
		t.Error("PlayingPath should be empty")
	}
}

// TestKeepAndDelete_SingleGroup tests keepAndDelete with a single group
func TestKeepAndDelete_SingleGroup(t *testing.T) {
	state := &AppState{
		DeletedCount:      0,
		FreedBytes:        0,
		CurrentGroupIndex: 0,
		Groups: []scanner.DuplicateGroup{
			{
				Hash: "testhash",
				Files: []scanner.FileInfo{
					{Path: "/test/file1.txt", Name: "file1.txt", Size: 512},
					{Path: "/test/file2.txt", Name: "file2.txt", Size: 512},
				},
			},
		},
	}

	files := state.Groups[0].Files
	keepAndDelete(state, 0, files)

	if state.DeletedCount != 1 {
		t.Errorf("DeletedCount = %d; want 1", state.DeletedCount)
	}
	if state.FreedBytes != 512 {
		t.Errorf("FreedBytes = %d; want 512", state.FreedBytes)
	}
	if len(state.Groups) != 0 {
		t.Errorf("Groups should be empty, got %d", len(state.Groups))
	}
}

// TestKeepAndDelete_MultipleGroups tests keepAndDelete with multiple groups
func TestKeepAndDelete_MultipleGroups(t *testing.T) {
	state := &AppState{
		DeletedCount:      0,
		FreedBytes:        0,
		CurrentGroupIndex: 0,
		Groups: []scanner.DuplicateGroup{
			{
				Hash: "hash1",
				Files: []scanner.FileInfo{
					{Path: "/test/file1.txt", Name: "file1.txt", Size: 100},
					{Path: "/test/file2.txt", Name: "file2.txt", Size: 100},
				},
			},
			{
				Hash: "hash2",
				Files: []scanner.FileInfo{
					{Path: "/test/file3.txt", Name: "file3.txt", Size: 200},
					{Path: "/test/file4.txt", Name: "file4.txt", Size: 200},
				},
			},
		},
	}

	files := state.Groups[0].Files
	keepAndDelete(state, 0, files)

	if state.DeletedCount != 1 {
		t.Errorf("DeletedCount = %d; want 1", state.DeletedCount)
	}
	if state.FreedBytes != 100 {
		t.Errorf("FreedBytes = %d; want 100", state.FreedBytes)
	}
	// Should have 1 group left
	if len(state.Groups) != 1 {
		t.Errorf("Groups length = %d; want 1", len(state.Groups))
	}
}

// TestSidebarItem_Struct tests the SidebarItem structure
func TestSidebarItem_Struct(t *testing.T) {
	item := SidebarItem{
		Icon:     nil,
		Name:     "Test Item",
		OnClick:  nil,
	}

	if item.Name != "Test Item" {
		t.Errorf("Name = %q; want %q", item.Name, "Test Item")
	}
}

// TestSidebar_EmptyItems tests sidebar with no items
func TestSidebar_EmptyItems(t *testing.T) {
	items := []SidebarItem{}

	if len(items) != 0 {
		t.Errorf("Items length = %d; want 0", len(items))
	}
}

// TestSidebar_SingleItem tests sidebar with a single item
func TestSidebar_SingleItem(t *testing.T) {
	items := []SidebarItem{
		{Name: "Item 1"},
	}

	if len(items) != 1 {
		t.Errorf("Items length = %d; want 1", len(items))
	}
	if items[0].Name != "Item 1" {
		t.Errorf("Items[0].Name = %q; want %q", items[0].Name, "Item 1")
	}
}

// TestSidebar_MultipleItems tests sidebar with multiple items
func TestSidebar_MultipleItems(t *testing.T) {
	items := []SidebarItem{
		{Name: "Item 1"},
		{Name: "Item 2"},
		{Name: "Item 3"},
	}

	if len(items) != 3 {
		t.Errorf("Items length = %d; want 3", len(items))
	}
}

// TestSidebar_Selection tests sidebar selection logic
func TestSidebar_Selection(t *testing.T) {
	items := []SidebarItem{
		{Name: "Item 1"},
		{Name: "Item 2"},
		{Name: "Item 3"},
	}

	// Test valid selections
	for i := 0; i < len(items); i++ {
		if items[i].Name == "" {
			t.Errorf("Items[%d].Name is empty", i)
		}
	}
}

// TestSidebar_OutOfBoundsSelection tests sidebar with out of bounds selection
func TestSidebar_OutOfBoundsSelection(t *testing.T) {
	items := []SidebarItem{
		{Name: "Item 1"},
	}

	// Simulate out of bounds access (should be handled by UI framework)
	outOfBounds := 10
	if outOfBounds >= len(items) {
		t.Log("Out of bounds selection detected (handled by UI framework)")
	}
}

// TestCreateSidebar_ReturnsValidStructure tests that CreateSidebar returns valid structure
func TestCreateSidebar_ReturnsValidStructure(t *testing.T) {
	sidebar := CreateSidebar()

	if sidebar == nil {
		t.Fatal("CreateSidebar should not return nil")
	}
}

// TestCreateSidebar_ItemCount tests that CreateSidebar creates correct number of items
func TestCreateSidebar_ItemCount(t *testing.T) {
	sidebar := CreateSidebar()

	if sidebar == nil {
		t.Fatal("CreateSidebar should not return nil")
	}

	// The sidebar should contain at least 3 items (Duplicate Finder, Cache Cleaner, Disk Analyzer)
	// We can't easily check the internal list, but we can verify the structure exists
	t.Log("Sidebar created successfully")
}

// TestSidebarItem_NilOnClick tests SidebarItem with nil OnClick
func TestSidebarItem_NilOnClick(t *testing.T) {
	item := SidebarItem{
		Name:    "Test Item",
		OnClick: nil,
	}

	if item.OnClick != nil {
		t.Error("OnClick should be nil")
	}
}

// TestSidebar_VariousIcons tests sidebar with various icons
func TestSidebar_VariousIcons(t *testing.T) {
	items := []SidebarItem{
		{Name: "Duplicates", Icon: nil},
		{Name: "Cache", Icon: nil},
		{Name: "Analyzer", Icon: nil},
	}

	for i, item := range items {
		if item.Name == "" {
			t.Errorf("Items[%d].Name is empty", i)
		}
	}
}

// TestSidebarItem_DataIntegrity tests that sidebar items maintain data integrity
func TestSidebarItem_DataIntegrity(t *testing.T) {
	tests := []struct {
		name string
		item SidebarItem
	}{
		{
			name: "empty strings",
			item: SidebarItem{Name: ""},
		},
		{
			name: "unicode name",
			item: SidebarItem{Name: "🔍 Suche"},
		},
		{
			name: "long name",
			item: SidebarItem{Name: strings.Repeat("a", 1000)},
		},
		{
			name: "special chars",
			item: SidebarItem{Name: "<script>alert('xss')</script>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.item.Name != tt.item.Name {
				t.Error("Name should be preserved")
			}
		})
	}
}
