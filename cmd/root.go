package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/PopolQue/dupclean/cli/interactive"
	"github.com/PopolQue/dupclean/internal/version"
	"github.com/PopolQue/dupclean/scanner"

	"github.com/spf13/cobra"
)

var (
	guiFlag        bool
	modeFlag       string
	allFlag        bool
	similarityFlag int
)

var rootCmd = &cobra.Command{
	Use:     "dupclean [folder]",
	Short:   "DupClean — Duplicate File Cleaner",
	Long:    `DupClean is a tool for finding duplicate files, analyzing disk space, and cleaning caches.`,
	Version: version.Version,
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if guiFlag {
			if LaunchGUI != nil {
				LaunchGUI()
				return nil
			}
			return fmt.Errorf("GUI mode is not available in this build.\nPlease download the full version with GUI from:\nhttps://github.com/PopolQue/github.com/PopolQue/dupclean/releases")
		}

		if len(args) == 0 {
			if LaunchGUI != nil {
				LaunchGUI()
				return nil
			}
			if err := cmd.Help(); err != nil {
				return err
			}
			return nil
		}

		folder := args[0]
		return runDuplicateFinder(cmd, folder)
	},
}

// LaunchGUI is a function that launches the GUI.
// It is set by the gui-enabled build.
var LaunchGUI func()

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&guiFlag, "gui", "g", false, "Launch GUI")
	rootCmd.Flags().StringVar(&modeFlag, "mode", "audio", "Scanner mode: audio, byte, photo")
	rootCmd.Flags().BoolVar(&allFlag, "all", false, "Scan all file types (same as --mode=byte)")
	rootCmd.Flags().IntVar(&similarityFlag, "similarity", 90, "Minimum similarity for photo mode (0-100)")

	// Set version template
	rootCmd.SetVersionTemplate("DupClean {{.Version}}\n")
}

var interactiveRun = interactive.Run

func runDuplicateFinder(cmd *cobra.Command, folder string) error {
	mode := modeFlag
	if allFlag {
		mode = "byte"
	}

	groups, stats, err := executeDuplicateFinder(cmd, folder, mode, similarityFlag)
	if err != nil {
		return err
	}

	interactiveRun(groups, stats)
	return nil
}

func executeDuplicateFinder(cmd *cobra.Command, folder string, mode string, similarity int) ([]scanner.DuplicateGroup, scanner.ScanStats, error) {
	folder = filepath.Clean(folder)
	absPath, err := filepath.Abs(folder)
	if err != nil {
		return nil, scanner.ScanStats{}, fmt.Errorf("invalid path '%s': %w", folder, err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, scanner.ScanStats{}, fmt.Errorf("cannot access '%s': %w", folder, err)
	}

	if !info.IsDir() {
		return nil, scanner.ScanStats{}, fmt.Errorf("'%s' is not a valid directory", folder)
	}

	sc, ok := scanner.GetScanner(mode)
	if !ok {
		return nil, scanner.ScanStats{}, fmt.Errorf("unknown mode '%s'. Available modes: %s", mode, strings.Join(scanner.AvailableModes(), ", "))
	}

	opts := scanner.Options{
		IncludeHidden: false,
		MinSize:       0,
		SimilarityPct: similarity,
		Context:       cmd.Context(),
	}

	return sc.Scan(absPath, opts)
}
