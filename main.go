//go:build !gui
// +build !gui

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"dupclean/scanner"
	"dupclean/ui"
)

const (
	flagGUI     = "--gui"
	flagGUIAlt  = "-g"
	flagHelp    = "--help"
	flagHelpAlt = "-h"
	flagAll     = "--all"
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
	fmt.Println("DupClean — Duplicate File Cleaner (CLI)")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  dupclean <folder>           Scan folder for duplicates")
	fmt.Println("  dupclean <folder> --all     Scan all files (not just audio)")
	fmt.Println("  dupclean --gui              Launch GUI (not available in CLI build)")
	fmt.Println("  dupclean --help             Show this help")
	fmt.Println()
	fmt.Println("Supported audio formats: .wav, .aiff, .aif, .mp3, .flac, .ogg, .m4a, .aac")
	fmt.Println()
	fmt.Println("Full version with GUI: https://github.com/PopolQue/dupclean/releases")
}
