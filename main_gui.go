//go:build gui
// +build gui

// Full build with GUI - used for GitHub Releases
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"dupclean/gui"
	"dupclean/scanner"
	"dupclean/ui"
)

const (
	flagGUI        = "--gui"
	flagGUIAlt     = "-g"
	flagHelp       = "--help"
	flagHelpAlt    = "-h"
	flagVersion    = "--version"
	flagVersionAlt = "-v"
	flagAll        = "--all"
)

func main() {
	if len(os.Args) >= 2 && (os.Args[1] == flagVersion || os.Args[1] == flagVersionAlt) {
		fmt.Printf("DupClean %s (GUI Build)\n", Version)
		os.Exit(0)
	}

	if len(os.Args) < 2 || os.Args[1] == flagGUI || os.Args[1] == flagGUIAlt {
		gui.RunGUI()
		return
	}

	if os.Args[1] == flagHelp || os.Args[1] == flagHelpAlt {
		printHelp()
		os.Exit(0)
	}

	folder := os.Args[1]
	scanAll := len(os.Args) > 2 && os.Args[2] == flagAll

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

	groups, stats, err := scanner.FindDuplicates(absPath, scanAll, nil, []string{}, []string{})
	if err != nil {
		fmt.Printf("Error: scan failed: %v\n", err)
		os.Exit(1)
	}

	ui.Run(groups, stats)
}

func printHelp() {
	fmt.Printf("DupClean %s — Duplicate File Cleaner\n", Version)
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  dupclean              Launch GUI")
	fmt.Println("  dupclean --gui        Launch GUI (same as above)")
	fmt.Println("  dupclean <folder>     Scan folder in CLI mode")
	fmt.Println("  dupclean <folder> --all   Scan all files (not just audio)")
	fmt.Println("  dupclean --version    Show version")
	fmt.Println()
	fmt.Println("Supported audio formats: .wav, .aiff, .aif, .mp3, .flac, .ogg, .m4a, .aac")
}
