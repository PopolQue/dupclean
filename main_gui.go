//go:build gui
// +build gui

// Full build with GUI - used for GitHub Releases
package main

import (
	"fmt"
	"os"

	"dupclean/gui"
)

const (
	flagGUI        = "--gui"
	flagGUIAlt     = "-g"
	flagHelp       = "--help"
	flagHelpAlt    = "-h"
	flagVersion    = "--version"
	flagVersionAlt = "-v"
	flagAll        = "--all"
	flagMode       = "--mode"
	flagSimilarity = "--similarity"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] == flagGUI || os.Args[1] == flagGUIAlt {
		gui.RunGUI()
		return
	}

	if os.Args[1] == flagHelp || os.Args[1] == flagHelpAlt {
		printHelp()
		os.Exit(0)
	}

	if os.Args[1] == flagVersion || os.Args[1] == flagVersionAlt {
		fmt.Printf("DupClean %s (GUI Build)\n", Version)
		os.Exit(0)
	}

	// Check for subcommands
	if os.Args[1] == "analyze" {
		runAnalyze(os.Args[2:])
		return
	}
	if os.Args[1] == "clean" {
		runClean(os.Args[2:])
		return
	}

	// CLI mode duplicate finder
	runDuplicateFinder(os.Args[1:])
}

func printHelp() {
	fmt.Printf("DupClean %s — Duplicate File Cleaner\n", Version)
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  dupclean              Launch GUI")
	fmt.Println("  dupclean --gui        Launch GUI")
	fmt.Println("  dupclean <folder>     Scan folder for duplicates")
	fmt.Println("  dupclean analyze <dir> Analyze disk usage")
	fmt.Println("  dupclean clean        Cleanup cache & temp files")
	fmt.Println("  dupclean --help       Show this help")
	fmt.Println("  dupclean --version    Show version")
	fmt.Println()
	fmt.Println("Duplicate Finder Options:")
	fmt.Println("  --mode=<mode>       Scanner mode: audio (default), byte, photo")
	fmt.Println("  --all               Scan all file types (same as --mode=byte)")
	fmt.Println("  --similarity=<pct>  Minimum similarity for photo mode (0-100, default: 90)")
	fmt.Println()
	fmt.Println("Disk Analyzer Options:")
	fmt.Println("  --top=N            Show N largest files (default: 20)")
	fmt.Println("  --depth=N          Tree depth in CLI view (default: 2)")
	fmt.Println("  --min-size=MB      Exclude files smaller than MB megabytes")
	fmt.Println("  --older-than=days  Only include files not modified in N days")
	fmt.Println("  --by-type          Show type breakdown instead of tree")
	fmt.Println("  --json             Output JSON to stdout")
	fmt.Println("  --no-hidden        Skip hidden files and folders")
	fmt.Println("  --follow-symlinks  Follow symbolic links")
	fmt.Println("  --exclude=pattern  Glob pattern to exclude (repeatable)")
	fmt.Println("  --workers=N        Number of concurrent stat workers")
	fmt.Println()
	fmt.Println("Cache Cleaner Options:")
	fmt.Println("  --category=CAT     system|browser|developer|logs")
	fmt.Println("  --target=ID        Only scan this specific target ID (repeatable)")
	fmt.Println("  --min-age=DURATION Only include files older than this (e.g. 24h, 7d)")
	fmt.Println("  --permanent        Delete permanently instead of moving to Trash")
	fmt.Println("  --dry-run          Show what would be deleted")
	fmt.Println("  --yes              Skip interactive confirmation")
	fmt.Println("  --workers=N        Number of concurrent scan workers")
}

func printAnalyzeHelp() {
	fmt.Println("DupClean — Disk Space Analyzer")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  dupclean analyze <folder> [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --top=N            Show N largest files (default: 20)")
	fmt.Println("  --depth=N          Tree depth in CLI view (default: 2)")
	fmt.Println("  --min-size=MB      Exclude files smaller than MB megabytes")
	fmt.Println("  --older-than=days  Only include files not modified in N days")
	fmt.Println("  --by-type          Show type breakdown instead of tree")
	fmt.Println("  --json             Output JSON to stdout")
	fmt.Println("  --no-hidden        Skip hidden files and folders")
	fmt.Println("  --follow-symlinks  Follow symbolic links")
	fmt.Println("  --exclude=pattern  Glob pattern to exclude (repeatable)")
	fmt.Println("  --workers=N        Number of concurrent stat workers")
}

func printCleanHelp() {
	fmt.Println("DupClean — Cache & Temp Cleaner")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  dupclean clean [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --category=CATEGORY   Only scan targets in this category (system|browser|developer|logs)")
	fmt.Println("  --target=ID           Only scan this specific target ID (repeatable)")
	fmt.Println("  --min-age=DURATION    Only include files older than this (e.g. 24h, 7d)")
	fmt.Println("  --permanent           Delete permanently instead of moving to Trash")
	fmt.Println("  --dry-run             Show what would be deleted without deleting anything")
	fmt.Println("  --yes                 Skip interactive confirmation")
	fmt.Println("  --workers=N           Number of concurrent scan workers")
}
