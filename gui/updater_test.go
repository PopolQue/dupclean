package gui

import (
	"testing"
)

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
