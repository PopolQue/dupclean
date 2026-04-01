package scanner

import (
	"testing"
)

func TestGetScanner_Audio(t *testing.T) {
	scanner, found := GetScanner("audio")

	if !found {
		t.Fatal("Expected audio scanner to be found")
	}
	if scanner == nil {
		t.Fatal("Expected non-nil audio scanner")
	}

	// Verify it's the right type
	if _, ok := scanner.(*AudioScanner); !ok {
		t.Error("Expected AudioScanner type")
	}
}

func TestGetScanner_Byte(t *testing.T) {
	scanner, found := GetScanner("byte")

	if !found {
		t.Fatal("Expected byte scanner to be found")
	}
	if scanner == nil {
		t.Fatal("Expected non-nil byte scanner")
	}

	// Verify it's the right type
	if _, ok := scanner.(*ByteScanner); !ok {
		t.Error("Expected ByteScanner type")
	}
}

func TestGetScanner_Photo(t *testing.T) {
	scanner, found := GetScanner("photo")

	if !found {
		t.Fatal("Expected photo scanner to be found")
	}
	if scanner == nil {
		t.Fatal("Expected non-nil photo scanner")
	}

	// Verify it's the right type
	if _, ok := scanner.(*PhotoScanner); !ok {
		t.Error("Expected PhotoScanner type")
	}
}

func TestGetScanner_Unknown(t *testing.T) {
	scanner, found := GetScanner("unknown")

	if found {
		t.Error("Expected false for unknown mode")
	}
	if scanner != nil {
		t.Error("Expected nil for unknown mode")
	}
}

func TestGetScanner_EmptyString(t *testing.T) {
	scanner, found := GetScanner("")

	if found {
		t.Error("Expected false for empty string mode")
	}
	if scanner != nil {
		t.Error("Expected nil for empty string mode")
	}
}

func TestGetScanner_CaseSensitive(t *testing.T) {
	// Test that mode names are case-sensitive
	scanner, found := GetScanner("AUDIO")

	if found {
		t.Error("Expected false for uppercase mode name")
	}
	if scanner != nil {
		t.Error("Expected nil for uppercase mode name")
	}

	_, found = GetScanner("Audio")
	if found {
		t.Error("Expected false for mixed case mode name")
	}
}

func TestAvailableModes(t *testing.T) {
	modes := AvailableModes()

	if len(modes) == 0 {
		t.Fatal("Expected at least one available mode")
	}

	// Check for expected modes
	foundAudio := false
	foundByte := false
	foundPhoto := false

	for _, mode := range modes {
		switch mode {
		case "audio":
			foundAudio = true
		case "byte":
			foundByte = true
		case "photo":
			foundPhoto = true
		}
	}

	if !foundAudio {
		t.Error("Expected 'audio' mode to be available")
	}
	if !foundByte {
		t.Error("Expected 'byte' mode to be available")
	}
	if !foundPhoto {
		t.Error("Expected 'photo' mode to be available")
	}
}

func TestAvailableModes_NoDuplicates(t *testing.T) {
	modes := AvailableModes()

	// Check for duplicates
	seen := make(map[string]bool)
	for _, mode := range modes {
		if seen[mode] {
			t.Errorf("Duplicate mode found: %q", mode)
		}
		seen[mode] = true
	}
}

func TestAvailableModes_Count(t *testing.T) {
	modes := AvailableModes()

	// Should have exactly 3 modes (audio, byte, photo)
	if len(modes) != 3 {
		t.Errorf("Expected 3 modes, got %d: %v", len(modes), modes)
	}
}

func TestScannerRegistry_Integrity(t *testing.T) {
	// Verify all registered scanners can be instantiated
	modes := AvailableModes()

	for _, mode := range modes {
		scanner, found := GetScanner(mode)
		if !found {
			t.Errorf("Mode %q not found in registry", mode)
		}
		if scanner == nil {
			t.Errorf("Mode %q returned nil scanner", mode)
		}
	}
}

func TestGetScanner_MultipleCalls(t *testing.T) {
	// Verify multiple calls return valid instances
	// (they may or may not be the same instance depending on implementation)
	scanner1, found1 := GetScanner("byte")
	scanner2, found2 := GetScanner("byte")

	if !found1 || !found2 {
		t.Fatal("Expected byte scanner to be found")
	}
	if scanner1 == nil || scanner2 == nil {
		t.Fatal("Expected non-nil scanners")
	}

	// Both should be ByteScanner type
	if _, ok := scanner1.(*ByteScanner); !ok {
		t.Error("Expected ByteScanner type for first call")
	}
	if _, ok := scanner2.(*ByteScanner); !ok {
		t.Error("Expected ByteScanner type for second call")
	}
}

func TestGetScanner_AudioIndependence(t *testing.T) {
	// Verify audio scanner calls return valid instances
	scanner1, found1 := GetScanner("audio")
	scanner2, found2 := GetScanner("audio")

	if !found1 || !found2 {
		t.Fatal("Expected audio scanner to be found")
	}
	if scanner1 == nil || scanner2 == nil {
		t.Fatal("Expected non-nil scanners")
	}
}

func TestGetScanner_PhotoIndependence(t *testing.T) {
	// Verify photo scanner calls return valid instances
	scanner1, found1 := GetScanner("photo")
	scanner2, found2 := GetScanner("photo")

	if !found1 || !found2 {
		t.Fatal("Expected photo scanner to be found")
	}
	if scanner1 == nil || scanner2 == nil {
		t.Fatal("Expected non-nil scanners")
	}
}

func TestAvailableModes_Stability(t *testing.T) {
	// Call AvailableModes multiple times and verify consistency
	modes1 := AvailableModes()
	modes2 := AvailableModes()

	if len(modes1) != len(modes2) {
		t.Errorf("Mode count changed: %d vs %d", len(modes1), len(modes2))
	}

	// Check same modes present
	modeSet := make(map[string]bool)
	for _, mode := range modes1 {
		modeSet[mode] = true
	}

	for _, mode := range modes2 {
		if !modeSet[mode] {
			t.Errorf("Mode %q missing from second call", mode)
		}
	}
}

func TestGetScanner_AllModes(t *testing.T) {
	// Test getting all registered modes
	expectedModes := []string{"audio", "byte", "photo"}

	for _, mode := range expectedModes {
		scanner, found := GetScanner(mode)
		if !found {
			t.Errorf("Mode %q should be available", mode)
		}
		if scanner == nil {
			t.Errorf("Mode %q returned nil", mode)
		}
	}
}
