package gui

import (
	"sync"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
)

// Mock AppState for testing
func newTestState() *AppState {
	return &AppState{
		Window: test.NewWindow(nil),
	}
}

func TestPlayFile_Concurrency(t *testing.T) {
	state := newTestState()

	// Use a WaitGroup to ensure all goroutines complete
	var wg sync.WaitGroup

	// Simulate rapid-fire calls
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// This needs a real file to not fail immediately in SafePlayMedia
			// But for concurrency testing, we can mock SafePlayMedia if needed.
			// Given the complexity of mocking, let's just trigger playFile
			// with a non-existent file to ensure it fails gracefully under race conditions
			playFile(state, "nonexistent.wav", nil)

			time.Sleep(time.Millisecond * 10)
			stopPlayback(state)
		}()
	}

	wg.Wait()
}
