package cleaner

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/PopolQue/dupclean/internal/trash"
)

// vars for testing to allow mocking OS behaviors
var (
	execCommandContext = exec.CommandContext
	goos               = runtime.GOOS
	userHomeDir        = os.UserHomeDir
	absPath            = filepath.Abs
	osRemoveAll        = os.RemoveAll
	moveToTrash        = trash.MoveToTrash
)
