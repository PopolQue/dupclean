//go:build !gui
// +build !gui

// CLI-only build (default) - used by Homebrew
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dupclean/scanner"
	"dupclean/ui"
)

const (
	flagGUI        = "--gui"
	flagGUIAlt     = "-g"
	flagHelp       = "--help"
	flagHelpAlt    = "-h"
	flagAll        = "--all"
	flagMode       = "--mode"
	flagSimilarity = "--similarity"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	if os.Args[1] == flagHelp || os.Args[1] == flagHelpAlt {
		printHelp()
		os.Exit(0)
	}

	if os.Args[1] == flagGUI || os.Args[1] == flagGUIAlt {
		fmt.Println("Error: GUI mode is not available in this build.")
		fmt.Println("Please download the full version with GUI from:")
		fmt.Println("https://github.com/PopolQue/dupclean/releases")
		os.Exit(1)
	}

	// Parse arguments
	folder := ""
	mode := "audio" // default mode
	similarity := 90

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if strings.HasPrefix(arg, "--") {
			switch {
			case arg == flagAll:
				mode = "byte"
			case strings.HasPrefix(arg, flagMode+"="):
				mode = strings.TrimPrefix(arg, flagMode+"=")
			case strings.HasPrefix(arg, flagSimilarity+"="):
				fmt.Sscanf(strings.TrimPrefix(arg, flagSimilarity+"="), "%d", &similarity)
			}
		} else if folder == "" {
			folder = arg
		}
	}

	if folder == "" {
		fmt.Println("Error: Please specify a folder to scan")
		printHelp()
		os.Exit(1)
	}

	// Validate and clean the path
	folder = filepath.Clean(folder)
	absPath, err := filepath.Abs(folder)
	if err != nil {
		fmt.Printf("Error: invalid path '%s': %v\n", folder, err)
		os.Exit(1)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Error: path '%s' does not exist\n", folder)
		} else if os.IsPermission(err) {
			fmt.Printf("Error: permission denied for '%s'\n", folder)
		} else {
			fmt.Printf("Error: cannot access '%s': %v\n", folder, err)
		}
		os.Exit(1)
	}

	if !info.IsDir() {
		fmt.Printf("Error: '%s' is not a valid directory\n", folder)
		os.Exit(1)
	}

	// Get scanner for mode
	sc, ok := scanner.GetScanner(mode)
	if !ok {
		fmt.Printf("Error: unknown mode '%s'\n", mode)
		fmt.Printf("Available modes: %s\n", strings.Join(scanner.AvailableModes(), ", "))
		os.Exit(1)
	}

	// Configure scanner options
	opts := scanner.Options{
		IncludeHidden:  false,
		MinSize:        0,
		SimilarityPct:  similarity,
		IgnoreFolders:  []string{},
		IgnoreExtensions: []string{},
	}

	groups, stats, err := sc.Scan(absPath, opts)
	if err != nil {
		fmt.Printf("Error: scan failed: %v\n", err)
		os.Exit(1)
	}

	ui.Run(groups, stats)
}

func printHelp() {
	fmt.Println("DupClean — Duplicate File Cleaner (CLI)")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  dupclean <folder> [options]     Scan folder for duplicates")
	fmt.Println("  dupclean --gui                  Launch GUI (not available in CLI build)")
	fmt.Println("  dupclean --help                 Show this help")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --mode=<mode>       Scanner mode: audio (default), byte, photo")
	fmt.Println("  --all               Scan all file types (same as --mode=byte)")
	fmt.Println("  --similarity=<pct>  Minimum similarity for photo mode (0-100, default: 90)")
	fmt.Println()
	fmt.Println("Modes:")
	fmt.Println("  audio   - Audio files only (.wav, .mp3, .flac, etc.)")
	fmt.Println("  byte    - All file types, exact byte-for-byte matches")
	fmt.Println("  photo   - Images only, finds similar (not just identical) photos")
	fmt.Println()
	fmt.Println("Supported audio formats: .wav, .aiff, .aif, .mp3, .flac, .ogg, .m4a, .aac")
	fmt.Println("Supported photo formats: .jpg, .png, .gif, .webp, .bmp, .tiff")
	fmt.Println()
	fmt.Println("Full version with GUI: https://github.com/PopolQue/dupclean/releases")
}
