package gui

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/PopolQue/dupclean/scanner"
	"github.com/PopolQue/dupclean/internal/trash"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// getLogFilePath returns a platform-appropriate path for the log file.
// It accepts a getEnv function and a pathSeparator to facilitate testing.
func getLogFilePath(getEnv func(string) string, pathSeparator rune) string {
	// Try TMPDIR environment variable (Unix, macOS)
	if tmpDir := getEnv("TMPDIR"); tmpDir != "" {
		return filepath.Join(tmpDir, "dupclean.log")
	}

	// Try TEMP environment variable (Windows)
	if tempDir := getEnv("TEMP"); tempDir != "" {
		return filepath.Join(tempDir, "dupclean.log")
	}

	// Try TMP environment variable (fallback)
	if tmpDir := getEnv("TMP"); tmpDir != "" {
		return filepath.Join(tmpDir, "dupclean.log")
	}

	// Platform-specific defaults
	switch pathSeparator {
	case '\\':
		// Windows default
		return filepath.Join(getEnv("USERPROFILE"), "AppData", "Local", "Temp", "dupclean.log")
	default:
		// Unix-like default
		return "/tmp/dupclean.log"
	}
}

// SetupLogging configures log output to a file and returns an error if setup fails.
func SetupLogging() error {
	logPath := getLogFilePath(os.Getenv, filepath.Separator)

	// Ensure directory exists
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	log.SetOutput(logFile)
	log.Println("DupClean starting...")
	log.Printf("Log file: %s", logPath)
	return nil
}

// ProcessManager handles global process state and button registration.
type ProcessManager struct {
	mu                     sync.Mutex
	isAnyProcessRunning    bool
	registeredStartButtons []*widget.Button
}

var moveToTrash = trash.MoveToTrash

// NewProcessManager creates a new instance.
func NewProcessManager() *ProcessManager {
	return &ProcessManager{
		registeredStartButtons: make([]*widget.Button, 0),
	}
}

// SetProcessRunning updates the process state and synchronizes all registered buttons.
func (pm *ProcessManager) SetProcessRunning(running bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.isAnyProcessRunning = running
	for _, btn := range pm.registeredStartButtons {
		if running {
			btn.Disable()
		} else {
			btn.Enable()
		}
	}
}

// RegisterStartButton registers a button for process state management.
func (pm *ProcessManager) RegisterStartButton(btn *widget.Button) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.registeredStartButtons = append(pm.registeredStartButtons, btn)
	if pm.isAnyProcessRunning {
		btn.Disable()
	}
}

type AppState struct {
	Window         fyne.Window
	ProcessManager *ProcessManager // Injected manager
	// ...

	ContentContainer   *fyne.Container // Reference to content area (preserves sidebar)
	FolderPath         string
	ScanAll            bool
	IncludeHidden      bool
	FollowSymlinks     bool
	SimilarityPct      int
	Depth              int
	IsScanning         bool
	ProgressText       string
	ProgressValue      float64
	Groups             []scanner.DuplicateGroup
	CurrentGroupIndex  int
	DeletedCount       int
	FreedBytes         int64
	SkippedCount       int
	SkippedFiles       []string
	Stats              scanner.ScanStats
	CurrentPlayer      *exec.Cmd
	StopPlayer         func()
	IgnoreFolders      []string
	IgnoreExtensions   []string
	PlayingPath        string
	progressComponents *progressComponents
	mu                 sync.RWMutex  // Protects concurrent access to state
	playerDone         chan struct{} // Signal when player goroutine is done
	RefreshResults     func()        // Callback to refresh the results UI
	Selections         [][]bool      // Tracks which files to keep in each group (matches Groups)
	CurrentMode        string        // current deduplication mode (byte, audio, photo)
}

type progressComponents struct {
	label  *widget.Label
	status *widget.Label
	bar    *widget.ProgressBar
}

// ... (App State)
