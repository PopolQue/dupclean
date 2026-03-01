package main

import (
	"fmt"
	"os"

	"dupclean/internal/scanner"
	"dupclean/internal/ui"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("🎧 DupClean — Audio Duplicate Cleaner")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  dupclean <folder>              Scan folder for duplicate audio files")
		fmt.Println("  dupclean <folder> --all        Also scan non-audio files")
		fmt.Println()
		fmt.Println("Supported audio formats: .wav, .aiff, .aif, .mp3, .flac, .ogg, .m4a, .aac")
		os.Exit(1)
	}

	folder := os.Args[1]
	scanAll := len(os.Args) > 2 && os.Args[2] == "--all"

	info, err := os.Stat(folder)
	if err != nil || !info.IsDir() {
		fmt.Printf("❌ Error: '%s' is not a valid directory\n", folder)
		os.Exit(1)
	}

	groups, stats, err := scanner.FindDuplicates(folder, scanAll)
	if err != nil {
		fmt.Printf("❌ Scan error: %v\n", err)
		os.Exit(1)
	}

	ui.Run(groups, stats)
}
