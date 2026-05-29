package gui

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"dupclean/scanner"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// getLogFilePath returns a platform-appropriate path for the log file.
func getLogFilePath() string {
	// Try TMPDIR environment variable (Unix, macOS)
	if tmpDir := os.Getenv("TMPDIR"); tmpDir != "" {
		return filepath.Join(tmpDir, "dupclean.log")
	}

	// Try TEMP environment variable (Windows)
	if tempDir := os.Getenv("TEMP"); tempDir != "" {
		return filepath.Join(tempDir, "dupclean.log")
	}

	// Try TMP environment variable (fallback)
	if tmpDir := os.Getenv("TMP"); tmpDir != "" {
		return filepath.Join(tmpDir, "dupclean.log")
	}

	// Platform-specific defaults
	switch filepath.Separator {
	case '\\':
		// Windows default
		return filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local", "Temp", "dupclean.log")
	default:
		// Unix-like default
		return "/tmp/dupclean.log"
	}
}

func init() {
	logPath := getLogFilePath()

	// Ensure directory exists
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		// Can't create log directory, skip logging to file
		log.Println("DupClean starting (no file logging)...")
		return
	}

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// Can't open log file, skip logging to file
		log.Println("DupClean starting (no file logging)...")
		return
	}

	log.SetOutput(logFile)
	log.Println("DupClean starting...")
	log.Printf("Log file: %s", logPath)
}

type AppState struct {
	Window             fyne.Window
	ContentContainer   *fyne.Container // Reference to content area (preserves sidebar)
	FolderPath         string
	ScanAll            bool
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

var (
	globalMu              sync.Mutex
	isAnyProcessRunning   bool
	registeredStartButtons []*widget.Button
)

// setProcessRunning updates the global process state and synchronizes all registered start buttons
func setProcessRunning(running bool) {
	globalMu.Lock()
	defer globalMu.Unlock()
	isAnyProcessRunning = running
	for _, btn := range registeredStartButtons {
		if running {
			btn.Disable()
		} else {
			// Button will be re-enabled by its own logic if ready (e.g. folder selected)
			// But for now, let's just trigger a refresh of its state if we have a way.
			// Actually, the individual tools should handle the "ready" part.
			// If we disable here, we might enable it when it's not ready.
			// So we need a more sophisticated way.
		}
	}
}

func registerStartButton(btn *widget.Button) {
	globalMu.Lock()
	defer globalMu.Unlock()
	registeredStartButtons = append(registeredStartButtons, btn)
	if isAnyProcessRunning {
		btn.Disable()
	}
}
