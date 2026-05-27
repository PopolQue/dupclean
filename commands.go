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

func runAnalyze(args []string) {
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
				val, err := strconv.Atoi(strings.TrimPrefix(arg, "--top="))
				if err != nil {
					fmt.Printf("Error: invalid value for --top: %v\n", err)
					os.Exit(1)
				}
				cliOpts.TopN = val
			case strings.HasPrefix(arg, "--depth="):
				val, err := strconv.Atoi(strings.TrimPrefix(arg, "--depth="))
				if err != nil {
					fmt.Printf("Error: invalid value for --depth: %v\n", err)
					os.Exit(1)
				}
				cliOpts.Depth = val
			case strings.HasPrefix(arg, "--older-than="):
				val, err := strconv.Atoi(strings.TrimPrefix(arg, "--older-than="))
				if err != nil {
					fmt.Printf("Error: invalid value for --older-than: %v\n", err)
					os.Exit(1)
				}
				cliOpts.OlderThan = val
			case strings.HasPrefix(arg, "--min-size="):
				sizeMB, err := strconv.Atoi(strings.TrimPrefix(arg, "--min-size="))
				if err != nil {
					fmt.Printf("Error: invalid value for --min-size: %v\n", err)
					os.Exit(1)
				}
				cliOpts.MinSize = int64(sizeMB) * 1024 * 1024
			case arg == "--no-hidden":
				opts.IncludeHidden = false
			case arg == "--follow-symlinks":
				opts.FollowSymlinks = true
			case strings.HasPrefix(arg, "--exclude="):
				opts.ExcludePaths = append(opts.ExcludePaths, strings.TrimPrefix(arg, "--exclude="))
			case strings.HasPrefix(arg, "--workers="):
				val, err := strconv.Atoi(strings.TrimPrefix(arg, "--workers="))
				if err != nil {
					fmt.Printf("Error: invalid value for --workers: %v\n", err)
					os.Exit(1)
				}
				opts.Concurrency = val
			case arg == "--help":
				printAnalyzeHelp()
				os.Exit(0)
			default:
				fmt.Printf("Error: unknown flag '%s'\n", arg)
				printAnalyzeHelp()
				os.Exit(1)
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
	folder := ""
	mode := "audio"
	similarity := 90

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--") {
			switch {
			case arg == flagAll:
				mode = "byte"
			case strings.HasPrefix(arg, flagMode+"="):
				mode = strings.TrimPrefix(arg, flagMode+"=")
			case strings.HasPrefix(arg, flagSimilarity+"="):
				val, err := strconv.Atoi(strings.TrimPrefix(arg, flagSimilarity+"="))
				if err != nil {
					fmt.Printf("Error: invalid value for --similarity: %v\n", err)
					os.Exit(1)
				}
				similarity = val
			default:
				fmt.Printf("Error: unknown flag '%s'\n", arg)
				printHelp()
				os.Exit(1)
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

func runClean(args []string) {
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
				duration, err := parseDuration(strings.TrimPrefix(arg, "--min-age="))
				if err != nil {
					fmt.Printf("Error: invalid value for --min-age: %v\n", err)
					os.Exit(1)
				}
				opts.MinAge = duration
			case strings.HasPrefix(arg, "--workers="):
				val, err := strconv.Atoi(strings.TrimPrefix(arg, "--workers="))
				if err != nil {
					fmt.Printf("Error: invalid value for --workers: %v\n", err)
					os.Exit(1)
				}
				opts.Concurrency = val
			case arg == "--help":
				printCleanHelp()
				os.Exit(0)
			default:
				fmt.Printf("Error: unknown flag '%s'\n", arg)
				printCleanHelp()
				os.Exit(1)
			}
		}
	}

	targets := cleaner.Registry()
	targets = cleaner.FilterTargets(targets, category, targetIDs, noDeveloper, noBrowser)

	if len(targets) == 0 {
		fmt.Println("No cleanable targets found for the specified filters.")
		return
	}

	fmt.Println("Scanning...")
	result, err := cleaner.Scan(targets, opts)
	if err != nil {
		fmt.Printf("Error during scan: %v\n", err)
		os.Exit(1)
	}

	cleaner.RenderCLI(result, cliOpts)
}

// parseDuration wraps time.ParseDuration to support days ('d')
func parseDuration(s string) (time.Duration, error) {
	if strings.HasSuffix(s, "d") {
		daysStr := strings.TrimSuffix(s, "d")
		days, err := strconv.Atoi(daysStr)
		if err != nil {
			return 0, err
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}
