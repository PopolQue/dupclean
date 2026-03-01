package main

import (
	"fmt"
	"os"

	"dupclean/gui"
	"dupclean/scanner"
	"dupclean/ui"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] == "--gui" || os.Args[1] == "-g" {
		gui.RunGUI()
		return
	}

	if os.Args[1] == "--help" || os.Args[1] == "-h" {
		printHelp()
		os.Exit(0)
	}

	folder := os.Args[1]
	scanAll := len(os.Args) > 2 && os.Args[2] == "--all"

	info, err := os.Stat(folder)
	if err != nil || !info.IsDir() {
		fmt.Printf("Error: '%s' is not a valid directory\n", folder)
		os.Exit(1)
	}

	groups, stats, err := scanner.FindDuplicates(folder, scanAll, nil)
	if err != nil {
		fmt.Printf("Scan error: %v\n", err)
		os.Exit(1)
	}

	ui.Run(groups, stats)
}

func printHelp() {
	fmt.Println("DupClean — Audio Duplicate Cleaner")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  dupclean              Launch GUI")
	fmt.Println("  dupclean --gui        Launch GUI (same as above)")
	fmt.Println("  dupclean <folder>     Scan folder in CLI mode")
	fmt.Println("  dupclean <folder> --all   Scan all files (not just audio)")
	fmt.Println()
	fmt.Println("Supported audio formats: .wav, .aiff, .aif, .mp3, .flac, .ogg, .m4a, .aac")
}
