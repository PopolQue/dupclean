package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dupclean/internal/version"
	"dupclean/scanner"
	"dupclean/ui"

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
	Run: func(cmd *cobra.Command, args []string) {
		if guiFlag {
			if LaunchGUI != nil {
				LaunchGUI()
				return
			}
			fmt.Println("Error: GUI mode is not available in this build.")
			fmt.Println("Please download the full version with GUI from:")
			fmt.Println("https://github.com/PopolQue/dupclean/releases")
			os.Exit(1)
		}

		if len(args) == 0 {
			if LaunchGUI != nil {
				LaunchGUI()
				return
			}
			_ = cmd.Help()
			return
		}

		folder := args[0]
		runDuplicateFinder(folder)
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

func runDuplicateFinder(folder string) {
	mode := modeFlag
	if allFlag {
		mode = "byte"
	}
	similarity := similarityFlag

	folder = filepath.Clean(folder)
	absPath, err := filepath.Abs(folder)
	if err != nil {
		fmt.Printf("Error: invalid path '%s': %v\n", folder, err)
		os.Exit(1)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		fmt.Printf("Error: cannot access '%s': %v\n", folder, err)
		os.Exit(1)
	}

	if !info.IsDir() {
		fmt.Printf("Error: '%s' is not a valid directory\n", folder)
		os.Exit(1)
	}

	sc, ok := scanner.GetScanner(mode)
	if !ok {
		fmt.Printf("Error: unknown mode '%s'\n", mode)
		fmt.Printf("Available modes: %s\n", strings.Join(scanner.AvailableModes(), ", "))
		os.Exit(1)
	}

	opts := scanner.Options{
		IncludeHidden: false,
		MinSize:       0,
		SimilarityPct: similarity,
	}

	groups, stats, err := sc.Scan(absPath, opts)
	if err != nil {
		fmt.Printf("Error: scan failed: %v\n", err)
		os.Exit(1)
	}

	ui.Run(groups, stats)
}
