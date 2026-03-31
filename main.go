//go:build !gui
// +build !gui

// CLI-only build (default) - used by Homebrew
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"dupclean/cleaner"
	"dupclean/diskanalyzer"
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

	// Check for subcommands
	if os.Args[1] == "analyze" {
		runAnalyze(os.Args[2:])
		return
	}
	if os.Args[1] == "clean" {
		runClean(os.Args[2:])
		return
	}

	// Legacy duplicate finder mode
	runDuplicateFinder(os.Args)
}

func runAnalyze(args []string) {
	// Parse analyze-specific flags
	root := ""
	opts := diskanalyzer.DefaultOptions()
	cliOpts := diskanalyzer.CLIOptions{
		Depth: 2,
	}
	jsonOutput := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--") {
			switch {
			case arg == "--json":
				jsonOutput = true
			case arg == "--by-type":
				cliOpts.ByType = true
			case strings.HasPrefix(arg, "--top="):
				cliOpts.TopN, _ = strconv.Atoi(strings.TrimPrefix(arg, "--top="))
			case strings.HasPrefix(arg, "--depth="):
				cliOpts.Depth, _ = strconv.Atoi(strings.TrimPrefix(arg, "--depth="))
			case strings.HasPrefix(arg, "--older-than="):
				cliOpts.OlderThan, _ = strconv.Atoi(strings.TrimPrefix(arg, "--older-than="))
			case strings.HasPrefix(arg, "--min-size="):
				sizeMB, _ := strconv.Atoi(strings.TrimPrefix(arg, "--min-size="))
				cliOpts.MinSize = int64(sizeMB) * 1024 * 1024
			case arg == "--no-hidden":
				opts.IncludeHidden = false
			case arg == "--follow-symlinks":
				opts.FollowSymlinks = true
			case strings.HasPrefix(arg, "--exclude="):
				opts.ExcludePaths = append(opts.ExcludePaths, strings.TrimPrefix(arg, "--exclude="))
			case strings.HasPrefix(arg, "--workers="):
				opts.Concurrency, _ = strconv.Atoi(strings.TrimPrefix(arg, "--workers="))
			case arg == "--help":
				printAnalyzeHelp()
				os.Exit(0)
			}
		} else if root == "" {
			root = arg
		}
	}

	if root == "" {
		fmt.Println("Error: Please specify a folder to analyze")
		printAnalyzeHelp()
		os.Exit(1)
	}

	// Validate path
	info, err := os.Stat(root)
	if err != nil {
		fmt.Printf("Error: cannot access '%s': %v\n", root, err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Printf("Error: '%s' is not a valid directory\n", root)
		os.Exit(1)
	}

	// Run analysis
	result, errors, err := diskanalyzer.Walk(root, opts)
	if err != nil {
		fmt.Printf("Error: analysis failed: %v\n", err)
		os.Exit(1)
	}

	// Print non-fatal errors to stderr
	for _, e := range errors {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", e)
	}

	// Output
	if jsonOutput {
		if err := diskanalyzer.ExportJSON(result, os.Stdout); err != nil {
			fmt.Printf("Error: JSON export failed: %v\n", err)
			os.Exit(1)
		}
	} else {
		diskanalyzer.RenderCLI(result, cliOpts)
	}
}

func runDuplicateFinder(args []string) {
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
				_, _ = fmt.Sscanf(strings.TrimPrefix(arg, flagSimilarity+"="), "%d", &similarity)
				// Note: error intentionally ignored, similarity stays at default if parsing fails
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
		IncludeHidden:    false,
		MinSize:          0,
		SimilarityPct:    similarity,
		IgnoreFolders:    []string{},
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
	fmt.Println("  dupclean analyze <folder> [opts]  Analyze disk usage")
	fmt.Println("  dupclean --gui                  Launch GUI (not available in CLI build)")
	fmt.Println("  dupclean --help                 Show this help")
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
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  dupclean analyze ~/Music")
	fmt.Println("  dupclean analyze ~/Documents --top=50 --by-type")
	fmt.Println("  dupclean analyze ~/Photos --older-than=365 --min-size=10")
	fmt.Println("  dupclean analyze / --json > disk-usage.json")
}

func runClean(args []string) {
	// Parse flags
	opts := cleaner.ScanOptions{}
	cliOpts := cleaner.CLIOptions{}
	var category string
	var targetIDs []string
	var noDeveloper, noBrowser bool

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--") {
			switch {
			case arg == "--dry-run":
				cliOpts.DryRun = true
			case arg == "--permanent":
				cliOpts.Permanent = true
			case arg == "--yes":
				cliOpts.Yes = true
			case arg == "--no-developer":
				noDeveloper = true
			case arg == "--no-browser":
				noBrowser = true
			case strings.HasPrefix(arg, "--category="):
				category = strings.TrimPrefix(arg, "--category=")
			case strings.HasPrefix(arg, "--target="):
				targetIDs = append(targetIDs, strings.TrimPrefix(arg, "--target="))
			case strings.HasPrefix(arg, "--min-age="):
				duration, _ := time.ParseDuration(strings.TrimPrefix(arg, "--min-age="))
				opts.MinAge = duration
			case strings.HasPrefix(arg, "--workers="):
				opts.Concurrency, _ = strconv.Atoi(strings.TrimPrefix(arg, "--workers="))
			case arg == "--help":
				printCleanHelp()
				os.Exit(0)
			}
		}
	}

	// Get targets
	targets := cleaner.Registry()
	targets = cleaner.FilterTargets(targets, category, targetIDs, noDeveloper, noBrowser)

	if len(targets) == 0 {
		fmt.Println("No cleanable targets found for the specified filters.")
		return
	}

	// Scan
	fmt.Println("Scanning...")
	result, err := cleaner.Scan(targets, opts)
	if err != nil {
		fmt.Printf("Error during scan: %v\n", err)
		os.Exit(1)
	}

	// Render CLI
	cleaner.RenderCLI(result, cliOpts)
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
	fmt.Println("  --yes                 Skip interactive confirmation (use with --target for scripting)")
	fmt.Println("  --no-developer        Exclude developer tool caches from scan")
	fmt.Println("  --no-browser          Exclude browser caches from scan")
	fmt.Println("  --workers=N           Number of concurrent workers (default: CPU count)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  dupclean clean")
	fmt.Println("  dupclean clean --category=system")
	fmt.Println("  dupclean clean --target=macos-user-cache --target=linux-user-cache")
	fmt.Println("  dupclean clean --dry-run")
	fmt.Println("  dupclean clean --permanent --yes --target=dev-xcode-derived")
}
