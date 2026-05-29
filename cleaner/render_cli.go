package cleaner

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"dupclean/internal/fsutil"
)

// CLIOptions configures the CLI renderer.
type CLIOptions struct {
	DryRun    bool
	Permanent bool
	Yes       bool // skip confirmation
}

// RenderCLI runs the interactive 3-stage CLI: scan → review → confirm.
func RenderCLI(result *ScanResult, opts CLIOptions) {
	reader := bufio.NewReader(os.Stdin)

	// Stage 1: Show scan results
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("  Scan Results")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Group by category
	categories := make(map[string][]*CleanTarget)
	for _, t := range result.Targets {
		if t.TotalSize > 0 {
			categories[t.Category] = append(categories[t.Category], t)
		}
	}

	// Sort categories
	catNames := make([]string, 0, len(categories))
	for cat := range categories {
		catNames = append(catNames, cat)
	}
	sort.Strings(catNames)

	for _, cat := range catNames {
		targets := categories[cat]
		fmt.Printf("  %s\n", cat)
		for _, t := range targets {
			riskIcon := "✓"
			switch t.Risk {
			case RiskModerate:
				riskIcon = "⚠"
			case RiskHigh:
				riskIcon = "✗"
			}

			sizeStr := fsutil.FormatBytes(t.TotalSize)
			fmt.Printf("  %s  %-30s  %s\n", riskIcon, t.Label, sizeStr)
		}
		fmt.Println()
	}

	fmt.Printf("  %s\n", strings.Repeat("─", 60))
	fmt.Printf("  Total reclaimable:  %s\n", fsutil.FormatBytes(result.TotalSize))
	fmt.Println()

	// Stage 2: Review and select
	fmt.Println("Select targets to clean. [space] toggle  [a] all  [n] none  [enter] confirm  [q] quit")
	fmt.Println()

	// Pre-select safe targets
	for _, t := range result.Targets {
		if t.TotalSize > 0 && (t.Risk == RiskSafe || t.Risk == RiskLow) {
			t.Selected = true
		}
	}

	for {
		visible := getVisibleTargets(result)
		printSelection(visible, result.TotalSize)

		fmt.Print("\n> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "", "enter":
			goto stage3
		case "q", "quit":
			fmt.Println("\nCancelled.")
			return
		case "a":
			selectTargets(result, true, false) // select safe/low only
		case "A":
			selectTargets(result, true, true) // select all including moderate
		case "n":
			selectTargets(result, false, false)
		default:
			// Try to parse as index to toggle
			if idx, err := strconv.Atoi(input); err == nil && idx > 0 && idx <= len(visible) {
				visible[idx-1].Selected = !visible[idx-1].Selected
			}
		}
	}

stage3:
	// Stage 3: Confirm and delete
	selected := getSelectedTargets(result)
	totalSize := getTotalSize(selected)

	if totalSize == 0 {
		fmt.Println("\nNo targets selected. Nothing to clean.")
		return
	}

	fmt.Println()
	fmt.Printf("Ready to clean %s across %d targets.\n", fsutil.FormatBytes(totalSize), len(selected))

	if opts.DryRun {
		fmt.Println("DRY RUN - No files will be deleted")
	} else if opts.Permanent {
		fmt.Println("Files will be PERMANENTLY deleted (not recoverable from Trash)")
	} else {
		fmt.Println("Files will be moved to Trash (recoverable)")
	}

	if !opts.Yes {
		fmt.Print("\n  [enter] Clean now   [d] dry run   [q] cancel\n\n> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "d":
			opts.DryRun = true
			fmt.Println("\nRunning dry run...")
		case "q", "quit":
			fmt.Println("\nCancelled.")
			return
		}
	}

	// Collect all entries to delete
	var allEntries []EntryInfo
	for _, t := range selected {
		allEntries = append(allEntries, t.Entries...)
	}

	// Delete
	fmt.Println()
	fmt.Println("Cleaning...")

	deleteOpts := DeleteOptions{
		DryRun:    opts.DryRun,
		Permanent: opts.Permanent,
		OnProgress: func(deleted int, freedBytes int64, current string) {
			fmt.Printf("\r  Processed %d files, freed %s...", deleted, fsutil.FormatBytes(freedBytes))
		},
	}

	deleteResult, err := Delete(allEntries, deleteOpts)
	if err != nil {
		fmt.Printf("\nError during deletion: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Println()
	fmt.Println("Done!")
	fmt.Printf("  Files deleted:  %d\n", deleteResult.Deleted)
	fmt.Printf("  Space freed:    %s\n", fsutil.FormatBytes(deleteResult.FreedBytes))
	if deleteResult.Skipped > 0 {
		fmt.Printf("  Skipped:        %d (files in use)\n", deleteResult.Skipped)
	}

	if !opts.DryRun && !opts.Permanent {
		fmt.Println()
		fmt.Println("Tip: Empty your Trash to reclaim disk space.")
	}
}

func printSelection(visibleTargets []*CleanTarget, totalReclaimable int64) {
	categories := make(map[string][]*CleanTarget)
	for _, t := range visibleTargets {
		categories[t.Category] = append(categories[t.Category], t)
	}

	catNames := make([]string, 0, len(categories))
	for cat := range categories {
		catNames = append(catNames, cat)
	}
	sort.Strings(catNames)

	i := 1
	for _, cat := range catNames {
		targets := categories[cat]
		fmt.Printf("  %s\n", cat)
		for _, t := range targets {
			checkbox := "[ ]"
			if t.Selected {
				checkbox = "[✓]"
			}
			sizeStr := fsutil.FormatBytes(t.TotalSize)
			fmt.Printf("  %2d. %s  %-30s  %s\n", i, checkbox, t.Label, sizeStr)
			i++
		}
		fmt.Println()
	}

	var selectedSize int64
	for _, t := range visibleTargets {
		if t.Selected {
			selectedSize += t.TotalSize
		}
	}
	fmt.Printf("  Selected: %s (out of %s total)\n", fsutil.FormatBytes(selectedSize), fsutil.FormatBytes(totalReclaimable))
}

func getVisibleTargets(result *ScanResult) []*CleanTarget {
	var visible []*CleanTarget
	for _, t := range result.Targets {
		if t.TotalSize > 0 {
			visible = append(visible, t)
		}
	}
	return visible
}

// selectTargets updates selection state based on risk levels.
func selectTargets(result *ScanResult, selected bool, includeModerate bool) {
	for _, t := range result.Targets {
		if t.TotalSize == 0 {
			continue
		}
		if selected {
			if includeModerate {
				if t.Risk <= RiskModerate {
					t.Selected = true
				}
			} else {
				if t.Risk <= RiskLow {
					t.Selected = true
				}
			}
		} else {
			t.Selected = false
		}
	}
}

// getSelectedTargets returns all selected targets.
func getSelectedTargets(result *ScanResult) []*CleanTarget {
	var selected []*CleanTarget
	for _, t := range result.Targets {
		if t.Selected && t.TotalSize > 0 {
			selected = append(selected, t)
		}
	}
	return selected
}

// getTotalSize returns the sum of TotalSize for all given targets.
func getTotalSize(targets []*CleanTarget) int64 {
	var total int64
	for _, t := range targets {
		total += t.TotalSize
	}
	return total
}
